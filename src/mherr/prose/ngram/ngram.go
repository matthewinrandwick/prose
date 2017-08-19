// Package ngram locates entries in a plain-text ngram flat file.
package ngram

import (
	"bytes"
	"fmt"
	"io"
	"mherr/prose/bsearch"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var ResourcePath = "../bin"

var Debug = false

const (
	maxRecLength    = 1024
	shortFragLength = 5
	maxFreq         = 1<<63 - 1
)

type Match struct {
	Text string
	Freq int
	Len  int
}

func (m Match) String() string {
	return fmt.Sprintf("text=%q freq=%v len=%v", m.Text, m.Freq, m.Len)
}

type record struct {
	Text []byte
	freq []byte
}

func newRecord(b []byte) *record {
	r := &record{}
	for i := 0; i < len(b); i++ {
		if b[i] == '\t' {
			r.Text = b[0:i]
			r.freq = b[i+1 : len(b)-1]
			break
		}
	}
	return r
}

func (r record) Freq() int {
	v, _ := strconv.ParseInt(string(r.freq), 10, 32)
	return int(v)
}

func (r record) String() string {
	return fmt.Sprintf("freq=%v text=%q", r.Freq(), string(r.Text))
}

func (r record) Len() int64 {
	// tab + \n
	return int64(len(r.freq) + 1 + len(r.Text) + 1)
}

var cfg = bsearch.Config{
	ChunkSize: 1024,
	Delimiter: '\n',
	Less: func(l, r []byte) bool {
		lrec := newRecord(l)
		return bytes.Compare(lrec.Text, r) < 0
	},
}

// Find returns the top n matches from the database.
func Find(filename, prefix string, length int) (Matches, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var ms matchArray

	// Find the last record earlier than the request. The next record will be >= our request.
	sought := []byte(prefix)
	r := bsearch.LowerBound(cfg, f, st.Size(), sought)
	if r.Err == io.EOF {
		return nil, nil
	}
	for {
		// Find the one after.
		nextOffset := r.End
		r = bsearch.Read(cfg, f, nextOffset)
		if r.Err == io.EOF {
			break
		}
		if r.Err != nil {
			return nil, r.Err
		}
		rec := newRecord(r.Data)
		if !bytes.HasPrefix(rec.Text, sought) {
			break
		}
		m := Match{string(rec.Text), rec.Freq(), length}
		ms.insert(m)
	}
	return ms.slice(), nil
}

// matchArray is an efficient data structure to store the top 5 hits.
type matchArray struct {
	matches    [5]Match
	lowest     int
	lowestFreq int
}

// insert adds a match to the array.
func (a *matchArray) insert(v Match) {
	if v.Freq <= a.lowestFreq {
		return
	}

	a.matches[a.lowest] = v

	// Find new lowest.
	a.lowest = 0
	a.lowestFreq = maxFreq
	for i, x := range a.matches {
		if x.Freq < a.lowestFreq {
			a.lowest = i
			a.lowestFreq = x.Freq
		}
	}
}

// slice returns a slice of matches.
func (a *matchArray) slice() Matches {
	var ms Matches
	for _, m := range a.matches {
		if m.Freq > 0 {
			ms = append(ms, m)
		}
	}
	sort.Sort(ms)
	return ms
}

type Matches []Match

func (ms Matches) Len() int { return len(ms) }
func (ms Matches) Less(i, j int) bool {
	if ms[i].Len > ms[j].Len {
		return true
	}
	if ms[i].Len < ms[j].Len {
		return false
	}
	return ms[i].Freq > ms[j].Freq
}
func (ms Matches) Swap(i, j int) {
	ms[i], ms[j] = ms[j], ms[i]
}

func (ms Matches) String() string {
	var buf bytes.Buffer
	buf.WriteString("\n")
	for i, m := range ms {
		buf.WriteString(fmt.Sprintf("%20q %7v\n", m.Text, m.Freq))
		if i > 5 {
			buf.WriteString(fmt.Sprintf("%20v\n", "..."))
			break
		}
	}
	return buf.String()
}

func Predictions(text string) (Matches, error) {
	line := strings.ToLower(text)
	ms, err := allMatches(line)
	if err != nil {
		return nil, err
	}
	for i := range ms {
		ms[i].Text = findSuffix(line, ms[i].Text)
	}

	if Debug {
		fmt.Printf("with suffix removed:%v\n", ms)
	}

	// Strip out duplicates.
	seen := make(map[string]bool)
	var res Matches
	for _, m := range ms {
		if m.Text == "" {
			continue
		}
		if seen[m.Text] {
			continue
		}
		seen[m.Text] = true
		res = append(res, m)
	}

	if Debug {
		fmt.Printf("with duplicates removed:%v", res)
	}

	return res, nil
}

func matchLastN(line, filename string, length int) (Matches, error) {
	var ms Matches
	words := strings.Split(line, " ")
	if len(words) > length {
		words = words[len(words)-length:]
	}
	joined := strings.Join(words, " ")

	// Empty string is expensive to look up.
	if strings.Trim(joined, " ") == "" {
		return nil, nil
	}

	rec, err := Find(filename, joined, length)
	if err != nil {
		return nil, err
	}
	if len(line) == 0 {
		rec = nil
	}
	ms = append(ms, rec...)

	if Debug {
		fmt.Printf("for %v %v-%v:%v", filename, length, ms)
	}

	return ms, nil
}

func allMatches(line string) (Matches, error) {
	line = strings.ToLower(line)

	fail := func(err error) (Matches, error) {
		return nil, fmt.Errorf("allMatches(%q): %v", line, err)
	}

	short := len(line) < shortFragLength
	files := []struct {
		filename string
		l        int
		short    bool
	}{
		{"ngrams.5.txt", 5, false},
		{"ngrams.4.txt", 4, false},
		{"ngrams.3.txt", 3, false},
		{"ngrams.2.txt", 2, false},
		{"ngrams.1.txt", 1, true},
		{"ngrams.1.all.txt", 1, false},
	}

	// Search all files in parallel.
	var wg sync.WaitGroup
	res := make([]matchRes, len(files))
	for index, file := range files {
		i := index
		f := file
		if short && !f.short {
			continue
		}
		wg.Add(1)
		go func() {
			res[i].ms, res[i].err = matchLastN(line, filepath.Join(ResourcePath, f.filename), f.l)
			wg.Done()
		}()
	}
	wg.Wait()

	// But combine all their results deterministically in the original order.
	var ms Matches
	for i, m := range res {
		if m.err != nil {
			return fail(fmt.Errorf("%v: %v", files[i].filename, m.err))
		}
		ms = append(ms, m.ms...)
	}

	if Debug {
		fmt.Printf("combined matches:%v", ms)
	}

	return ms, nil
}

type matchRes struct {
	ms  Matches
	err error
}

func findSuffix(line, patch string) string {
	for e := range line {
		l := line[e:]
		if strings.HasPrefix(patch, l) {

			// Only get the next word.
			// Preserve leading space.
			l = strings.TrimPrefix(patch, l)
			if len(l) > 2 {
				i := strings.IndexRune(l[1:], ' ')
				if i != -1 {
					l = l[:i+1]
				}
			}
			return l
		}
	}
	return ""
}
