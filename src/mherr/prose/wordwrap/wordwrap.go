// Package wordwrap implements wordwrap.
package wordwrap

// WrapString wraps the given string within lim width in characters.
import (
	"bytes"
	"strings"
)

type parState int

const (
	outPar parState = iota
	prePar
	normPar
)

const spaces = " \t"

// Fold breaks the input into paragraphs with lines no longer than lim runes.
func Fold(inp string, lim int) []string {
	var res []string
	var line bytes.Buffer
	lastSpace := 0
	inPar := false
	lastPre := false

	emit := func(endPar bool) {
		s := line.String()
		isPre := preformatted(s)
		if !isPre && lastPre {
			res = append(res, "")
		}
		res = append(res, s)
		if !isPre && endPar {
			res = append(res, "")
		}
		lastPre = isPre
		inPar = false
		line.Reset()
		lastSpace = 0
	}

	for _, c := range inp {
		if !inPar {
			if c != '\n' {
				line.WriteRune(c)
				inPar = true
			}
		} else {
			switch {
			case c == ' ':
				line.WriteRune(c)
				lastSpace = line.Len()
			case c == '\n':
				emit(true)
			default:
				line.WriteRune(c)
			}

			if line.Len() > lim {
				if lastSpace == 0 {
					lastSpace = line.Len()
				}
				head := line.String()[:lastSpace]
				tail := line.String()[lastSpace:]
				head = strings.TrimRight(head, spaces)
				line.Reset()
				line.WriteString(head)
				emit(false)

				line.Reset()
				line.WriteString(tail)
				lastSpace = 0
			}
		}
	}

	if line.Len() != 0 {
		emit(false)
	}

	// Trim trailing blank lines.
	last := len(res)
	for i := len(res) - 1; i >= 0; i-- {
		if len(res[i]) == 0 {
			last = i
		} else {
			break
		}
	}
	return res[:last]
}

func preformatted(l string) bool {
	return len(l) > 0 && (l[0] == ' ' || l[0] == '\t')
}

// Unfold converts word-wrapped paragraphs back into unfolded line-per-paragraph input.
func Unfold(lines []string) []byte {
	if len(lines) == 0 {
		return nil
	}

	var (
		pars []string
		st   = outPar
		par  bytes.Buffer
	)

	emit := func() {
		if par.Len() == 0 {
			return
		}
		pars = append(pars, par.String())
		par.Reset()
	}

	for _, l := range lines {
		var newSt parState
		switch {
		case l == "":
			newSt = outPar
		case preformatted(l):
			newSt = prePar
		default:
			newSt = normPar
		}

		switch {
		case st == normPar && newSt == outPar:
			emit()
		case st == normPar && newSt == normPar:
			par.WriteRune(' ')
			par.WriteString(l)
		case st == normPar && newSt == prePar:
			emit()
			par.WriteString(l)
		case st == outPar && newSt == outPar:
			// Do nothing
		case st == outPar && newSt == normPar:
			par.WriteString(l)
		case st == outPar && newSt == prePar:
			par.WriteString(l)
		case st == prePar && newSt == outPar:
			emit()
		case st == prePar && newSt == normPar:
			emit()
			par.WriteString(l)
		case st == prePar && newSt == prePar:
			par.WriteRune('\n')
			par.WriteString(l)
		}
		st = newSt
	}
	emit()

	var buf bytes.Buffer
	for i, p := range pars {
		buf.WriteString(p)
		buf.WriteRune('\n')
		if i == len(pars)-1 {
			break
		}
		buf.WriteRune('\n')
	}
	return buf.Bytes()
}
