package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"arbovm/levenshtein"
)

var punct = map[byte]bool{
	'\\': true,
	'`':  true,
	'"':  true,
	'\n': true,
	' ':  true,
	'\r': true,
	',':  true,
	'.':  true,
	'!':  true,
	'[':  true,
	']':  true,
	'^':  true,
	'(':  true,
	')':  true,
	'?':  true,
	'_':  true,
}

var (
	from = flag.String("from", "", "The dictionary file that is the source dialect.")
	to   = flag.String("to", "", "The dictionary file for the target dialect.")
)

func usage() {
	flag.Usage()
	os.Exit(2)
}

type lang struct {
	m   map[string]int
	pre map[string]int
	s   []string
}

func load(filename string) (*lang, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	l := &lang{
		m:   make(map[string]int),
		pre: make(map[string]int),
	}
	s := bufio.NewScanner(f)
	line := 0
	for s.Scan() {
		t := s.Text()
		l.m[t] = line
		for x := range t {
			i := len(t) - x
			p := t[:i]
			if _, ok := l.pre[p]; ok {
				break
			}
			l.pre[t[:i]] = line
		}
		l.s = append(l.s, t)
		line++
	}

	return l, nil
}

func main() {
	flag.Usage = func() { fmt.Print("usage: trans-english  [filename]...\n") }
	flag.Parse()

	filenames := flag.Args()
	if len(filenames) == 0 {
		usage()
	}
	if *to == "" {
		usage()
	}
	if *from == "" {
		usage()
	}

	fail := func(err error) {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	src, err := load(*from)
	if err != nil {
		fail(err)
	}
	dst, err := load(*to)
	if err != nil {
		fail(err)
	}

	srcDst := makeLookup(src, dst)
	_ = srcDst

	for _, f := range filenames {
		if err := process(f, srcDst); err != nil {
			fail(err)
		}
	}
}

func process(filename string, m map[string]string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	s := bufio.NewScanner(f)
	for s.Scan() {
		l := s.Text() + "\n"
		st := 0
	line:
		for i := range l {
			c := l[i]
			switch {
			case i == len(l)-1:
				fallthrough
			case punct[c]:
			default:
				continue line
			}

			w := l[st:i]

			flush := func() {
				out.WriteString(w)
				out.WriteByte(c)
				st = i + 1
			}

			if len(w) == 0 {
				flush()
				continue
			}

			// First look up the word in the dictionary, as-is.
			nw, ok := m[w]
			if ok {
				w = nw
				flush()
				continue
			}

			// Convert upper-case words.
			if w == strings.ToUpper(w) {
				w2 := strings.ToLower(w)
				nw, ok := m[w2]
				if ok {
					w = strings.ToUpper(nw)
					flush()
					continue
				}
			}

			// Convert title-case words.
			if w == strings.Title(w) {
				w2 := strings.ToLower(w)
				nw, ok := m[w2]
				if ok {
					w = strings.Title(nw)
					flush()
					continue
				}
			}

			flush()
		}
	}

	return nil
}

func makeLookup(src, dst *lang) map[string]string {
	res := make(map[string]string)
	for s := range src.m {
		d := closest(s, dst)
		if d == s {
			continue
		}
		res[s] = d
	}
	return res
}

func closest(s string, dst *lang) string {
	if _, ok := dst.m[s]; ok {
		return s
	}

	pre := ""
	prei := 0
	for x := range s {
		i := len(s) - x
		p := s[:i]
		if pi, ok := dst.pre[p]; ok {
			pre = p
			prei = pi
			break
		}
	}
	if pre == "" {
		return ""
	}

	i := prei
	min := ""
	mind := -1
	for {
		if i >= len(dst.s) {
			break
		}
		d := dst.s[i]
		if !strings.HasPrefix(d, pre) {
			break
		}
		dis := levenshtein.Distance(s, d)
		if mind == -1 || dis < mind {
			min = d
			mind = dis
		}
		i++
	}
	return min
}
