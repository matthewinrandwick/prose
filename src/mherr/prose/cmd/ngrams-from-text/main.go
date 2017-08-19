// Makes ngrams from plain text files.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

var (
	errExit = errors.New("exit requested")

	punct = map[byte]bool{
		'\\': true,
		'`':  true,
		'"':  true,
		'\n': true,
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

	// Punctuation which may appear within words as well as without.
	ambiguous = map[byte]bool{
		'\'': true,
		'-':  true,
	}

	filter = flag.Int("filter", 0, "If specified, only ngrams of this length will be emitted.")
)

func usage() {
	flag.Usage()
	os.Exit(2)
}

func main() {
	flag.Usage = func() { fmt.Print("usage: make-ngrams -filter [1-5] [filename]...\n") }
	flag.Parse()

	filenames := flag.Args()
	if len(filenames) == 0 {
		usage()
	}
	if !(*filter >= 0 && *filter <= 5) {
		usage()
	}
	maxNgrams := *filter
	if maxNgrams == 0 {
		maxNgrams = 5
	}

	fail := func(err error) {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	ngrams := make(map[string]int)
	for _, fname := range filenames {
		f, err := os.Open(fname)
		if err != nil {
			fail(err)
		}
		fmt.Fprintf(os.Stderr, ".")
		err = process(f, ngrams, maxNgrams)
		f.Close()
		if err != nil && err != io.EOF {
			fail(err)
		}
	}
	fmt.Fprintf(os.Stderr, "\n")

	var words []string
	for s := range ngrams {
		words = append(words, s)
	}
	sort.Strings(words)

	for _, s := range words {
		fmt.Printf("%v\t%v\n", s, ngrams[s])
	}
}

type parseState int

const (
	beforeLine parseState = iota
	inWord
	outWord
)

func process(f *os.File, ngrams map[string]int, maxNgrams int) error {
	var (
		chunk   [4096]byte
		state   = beforeLine
		line    bytes.Buffer
		offsets = make([]int, maxNgrams)
		ws      = 0

		emit = func() {
			l := line.String()
			if len(l) == 0 {
				return
			}
			l = l[:len(l)-1]
			off := 0
			w := 0
			for {
				v := l[off:]
				if *filter > 0 {
					if strings.Count(v, " ") != maxNgrams-1 {
						break
					}
				}

				ngrams[v]++
				w++
				if w >= ws-1 {
					break
				}
				off = offsets[w]
			}
		}

		pop = func() bool {
			if ws == 0 {
				return false
			}
			diff := offsets[0]
			l := line.Bytes()[diff:]
			line.Reset()
			line.Write(l)
			for i := 0; i < ws-1; i++ {
				offsets[i] = offsets[i+1] - diff
			}
			ws--
			offsets[maxNgrams-1] = 0
			return true
		}

		recordLine = func() {
			for {
				emit()
				if !pop() {
					break
				}
			}
			line.Reset()
		}

		recordLetter = func(c byte) {
			line.WriteByte(c)
		}

		markWordEnd = func() {
			line.WriteByte(' ')
			offsets[ws] = line.Len()
			ws++

			if ws < maxNgrams {
				return
			}

			// Write out one entry.
			emit()
			pop()
		}
	)

	for {
		n, err := f.Read(chunk[:])
		ch := chunk[:n]
		for _, c := range ch {
			switch {
			// Some punctuation like ' can appear within a word, or without.
			case punct[c] || (ambiguous[c] && state != inWord):
				switch state {
				case beforeLine:
				case inWord:
					markWordEnd()
					recordLine()
				case outWord:
					recordLine()
				}
				state = beforeLine
			case c == '\t' || c == ' ':
				switch state {
				case beforeLine:
				case outWord:
				case inWord:
					markWordEnd()
					state = outWord
				}
			case c >= 'A' && c <= 'z' || c == '\'':
				// Convert to lowercase.
				if c >= 'A' && c <= 'Z' {
					c += 32
				}
				recordLetter(c)
				state = inWord
			}
		}
		if err != nil {
			return err
		}
	}

	return nil
}
