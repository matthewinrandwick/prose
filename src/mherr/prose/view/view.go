// Package view manages the rendering of a text document in the console.

package view

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mherr/prose/conio"
	"mherr/prose/ngram"
	"mherr/prose/wordwrap"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Doc is an in-memory document.
type Doc struct {
	filename      string
	auto          bool
	dirty         bool
	lastDraw      map[int]string
	lines         []string
	viewX, viewY  int
	y, x          int
	width, height int
	predictions   ngram.Matches
}

// New creates a new document from the given file.
func New(filename string) (*Doc, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	w, h, err := conio.Size()
	if err != nil {
		return nil, err
	}

	d := &Doc{
		filename: filename,
		width:    w,
		height:   h,
		lastDraw: make(map[int]string),
		auto:     true,
	}
	d.lines = wordwrap.Fold(string(data), d.textWidth())
	if len(d.lines) == 0 {
		d.lines = []string{""}
	}

	return d, nil
}

func (d *Doc) WindowChanged() error {
	d.lastDraw = make(map[int]string)
	w, h, err := conio.Size()
	if err != nil {
		return err
	}
	d.height = h
	d.width = w
	d.Redraw()
	d.hidePredictions()
	return nil
}

func (d *Doc) Redraw() {
	d.trimView()
	for y := 0; y < d.textHeight(); y++ {
		var l string
		p := d.viewY + y
		if p < len(d.lines) {
			l = d.lines[p]
			if p == d.y {
				l = l[d.viewX:]
				if d.viewX > 0 {
					l = "<" + l[1:]
				}
			}
			if len(l) > d.textWidth() {
				l = l[:d.textWidth()] + ">"
			}
		}
		d.drawLine(y+1, l)
	}
	d.drawStatusLine()
	d.moveCursor()
}

func (d *Doc) drawLine(y int, l string) {
	last, ok := d.lastDraw[y]
	if !ok || last != l {
		conio.Pos(y, 1)
		conio.Escape(conio.ClearLine) // Erase entire line
		d.lastDraw[y] = l
		conio.Out(l)
	}
}

func (d *Doc) Dirty() bool {
	return d.dirty
}

func (d *Doc) drawStatusLine() {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%5v:%3v ", d.y+1, d.x+1)
	b.WriteString(" | ")
	b.WriteString("[Ctl-S]ave")
	b.WriteString(" [Ctl-D]one")
	b.WriteString(" [Ctl-A]uto")
	b.WriteString(" [Ctl-O]ff")
	b.WriteString(" | ")
	if d.dirty {
		b.WriteString("*")
	}
	b.WriteString(filepath.Base(d.filename))
	d.WriteStatus(b.String())
}

func (d *Doc) WriteStatus(s string) {
	conio.Pos(d.statusBarY(), 1)
	conio.Escape("0;37m") // White text
	conio.Escape("40m")   // Grey backgrtound
	conio.Out(strings.Repeat(" ", d.Width()))
	conio.Pos(d.statusBarY(), 1)
	if len(s) > d.textWidth() {
		s = s[:d.textWidth()]
	}
	conio.Out(s)
	conio.Escape("0m") // Reverse video off
}

func (d *Doc) moveCursor() {
	conio.Pos(d.y-d.viewY+1, d.x-d.viewX+1)
}

func (d *Doc) Auto(auto bool) {
	d.auto = auto
	d.lastDraw = make(map[int]string)
	d.Redraw()
	d.showPredictions()
}

func (d *Doc) checkBounds() {
	if d.y < 0 {
		panic(fmt.Sprintf("d.y=%v", d.y))
	}
	if d.x < 0 {
		panic(fmt.Sprintf("d.x=%v", d.x))
	}
}

func (d *Doc) Move(dy, dx int) {
	defer d.hidePredictions()
	bounce := false
	i := 0
	for {
		d.y += dy
		d.x += dx
		i++
		if i > 10 {
			d.WriteStatus(fmt.Sprintf("d.x=%v d.y=%v bounce=%v dx=%v dy=%v", d.x, d.y, bounce, dx, dy))
			return
		}
		if d.y < 0 {
			d.y = 0
			bounce = true
		}
		if l := len(d.lines); d.y >= l {
			d.y = l - 1
			if d.y < 0 {
				d.y = 0
				bounce = true
			}
		}
		if d.x < 0 {
			d.x = 0
		}
		if l := len(d.lines[d.y]); d.x > l {
			d.x = l
		}
		if bounce {
			break
		}
		if len(d.lines[d.y]) != 0 {
			break
		}
	}
	d.checkBounds()
	d.trimView()
	d.Redraw()
}

func (d *Doc) trimView() {
	if d.y >= d.viewY+d.textHeight() {
		d.viewY = d.y - d.textHeight() + 1
	}
	if d.y < d.viewY {
		d.viewY = d.y
	}
	if d.x > d.viewX+d.textWidth() {
		d.viewX = d.x - d.textWidth() + 1
	}
	if d.x < d.viewX {
		d.viewX = d.x
	}
	if len(d.lines) == 0 {
		d.viewX = 0
		return
	}
	if len(d.lines[d.y]) < d.textWidth() {
		d.viewX = 0
	}
}

func (d *Doc) Delete() {
	d.dirty = true
	here := d.lines[d.y]
	if d.x == len(here) {
		if d.y == len(d.lines)-1 {
			return
		}

		out := append([]string{}, d.lines[:d.y]...)
		out = append(out, d.lines[d.y]+d.lines[d.y+1])
		out = append(out, d.lines[d.y+2:]...)
		d.lines = out
		d.Redraw()
		return
	}
	d.lines[d.y] = here[:d.x] + here[d.x+1:]
	d.deleteReflow()
	d.Redraw()
}

func (d *Doc) Enter() {
	d.dirty = true

	count := 2
	if l := d.lines[d.y]; l != "" && l[0] == ' ' {
		count = 1
	}

	// Add two lines, to create a new paragraph.
	for i := 0; i < count; i++ {
		var here, before, after string
		here = d.lines[d.y]
		if len(here) > 0 {
			before = here[:d.x]
			after = here[d.x:]
		}
		out := append([]string{}, d.lines[:d.y]...)
		out = append(out, before, after)
		out = append(out, d.lines[d.y+1:]...)
		d.lines = out
		d.y++
		d.x = 0
	}
	d.deleteReflow()
	d.Redraw()
	d.showPredictions()
}

func (d *Doc) char() byte {
	if d.y >= len(d.lines) {
		return 0
	}
	x := d.x
	if l := len(d.lines[d.y]); x >= l {
		x = l - 1
	}
	if x <= 0 {
		return 0
	}
	return d.lines[d.y][x]
}

func (d *Doc) CtlBackspace() {
	for {
		d.backspace()
		if d.x == 0 || d.char() == ' ' {
			break
		}
	}
	d.deleteReflow()
	d.Redraw()
}

func (d *Doc) Backspace() {
	d.backspace()
	d.deleteReflow()
	d.Redraw()
	d.hidePredictions()
}

func (d *Doc) backspace() {
	d.dirty = true
	if d.x == 0 && d.y == 0 {
		return
	}
	if d.x == 0 {
		moved := d.lines[d.y]
		d.lines[d.y-1] += moved
		out := append([]string{}, d.lines[:d.y]...)
		out = append(out, d.lines[d.y+1:]...)
		d.lines = out
		d.y--
		d.x = len(d.lines[d.y]) - len(moved)
	} else {
		d.lines[d.y] = d.lines[d.y][:d.x-1] + d.lines[d.y][d.x:]
		d.x--
	}
}

func (d *Doc) Edit(b byte) error {
	d.dirty = true
	var here, before, after string
	here = d.lines[d.y]
	if len(here) > 0 {
		before = here[:d.x]
		after = here[d.x:]
	}

	switch {
	case d.auto && b == '\t':
		fallthrough
	case d.auto && b == ';':
		d.addPrediction(0)
	case d.auto && b >= '1' && b <= '7':
		d.addPrediction(int(b) - '0')

		// Delete spaces before punctuation.
	case d.auto && b == ',':
		fallthrough
	case d.auto && b == '?':
		fallthrough
	case d.auto && b == '.':
		b := strings.TrimRight(before, " ")
		diff := len(before) - len(b)
		before = b
		d.x -= diff
		fallthrough

	default:
		d.lines[d.y] = before + string(b) + after
		d.x++
	}

	d.reflow()
	d.Redraw()
	return d.showPredictions()
}

func (d *Doc) Save() error {
	d.dirty = false
	if err := ioutil.WriteFile(d.filename+".tmp", wordwrap.Unfold(d.lines), 0666); err != nil {
		return err
	}
	if err := os.Rename(d.filename+".tmp", d.filename); err != nil {
		return err
	}

	d.Redraw()
	return nil
}

func (d *Doc) addPrediction(i int) {
	if i >= len(d.predictions) {
		return
	}
	word := d.predictions[i].Text + " "

	var here, before, after string
	here = d.lines[d.y]
	if len(here) > 0 {
		before = here[:d.x]
		after = here[d.x:]
	}
	d.lines[d.y] = before + word + after
	d.x += len(word)
}

func (d *Doc) debug(pat string, args ...interface{}) {
	conio.Escapef(conio.Home)
	conio.Outf(pat, args...)
	d.moveCursor()
}

func predictions(line string) (ngram.Matches, error) {
	res, err := ngram.Predictions(line)
	if err != nil {
		return nil, err
	}

	space, err := ngram.Predictions(line + " ")
	if err != nil {
		return nil, err
	}
	for _, p := range space {
		p.Text = " " + p.Text
		res = append(res, p)
	}
	sort.Sort(res)
	return res, nil
}

func (d *Doc) showPredictions() error {
	if !d.auto {
		d.hidePredictions()
		return nil
	}
	line := d.lines[d.y][:d.x]

	var err error
	d.predictions, err = predictions(line)
	if err != nil {
		return err
	}
	ws := strings.Split(line, " ")
	lastWord := ws[len(ws)-1]

	for i := 0; i < d.predictionsHeight(); i++ {
		var s string
		if i < len(d.predictions) {
			m := d.predictions[i]
			s = fmt.Sprintf("%v %v%v", string(shortcuts[i]), lastWord, m.Text)
		}
		d.drawLine(d.predictionsY()+i, s)
	}
	d.moveCursor()
	return nil
}

func (d *Doc) hidePredictions() {
	for i := 0; i < d.predictionsHeight(); i++ {
		d.drawLine(d.predictionsY()+i, "")
	}
	d.moveCursor()
}

const shortcuts = ";123456789"

// deleteReflow combines the following line with the current line, then reflows.
func (d *Doc) deleteReflow() {
	if d.y >= len(d.lines)-1 {
		return
	}

	if d.lines[d.y+1] == "" {
		return
	}

	out := append([]string{}, d.lines[:d.y]...)
	out = append(out, d.lines[d.y]+" "+d.lines[d.y+1])
	out = append(out, d.lines[d.y+2:]...)
	d.lines = out

	d.reflow()
}

func (d *Doc) reflow() {
	var carry string
	y := d.y
	for {
		if y >= len(d.lines) {
			if carry != "" {
				d.lines = append(d.lines, carry)
			}
			break
		}

		if carry != "" {
			if d.lines[y] == "" {
				tmp := append([]string{}, d.lines[:y]...)
				tmp = append(tmp, "")
				tmp = append(tmp, d.lines[y:]...)
				d.lines = tmp
			}
			d.lines[y] = strings.TrimSuffix(carry+" "+d.lines[y], " ")
		}

		if len(d.lines[y]) < d.textWidth() {
			break
		}

		for x := d.textWidth() - 1; x >= 0; x-- {
			if d.lines[y][x] == ' ' {
				carry = d.lines[y][x+1:]
				d.lines[y] = d.lines[y][:x]
				y++
				break
			}
		}

		if carry == "" {
			break
		}
	}

	diff := d.x - len(d.lines[d.y])
	if diff > 0 {
		d.y++
		d.x = diff - 1
	}

	if d.y >= len(d.lines) {
		d.lines = append(d.lines, "")
	}
	if l := len(d.lines[d.y]); d.x > l {
		d.x = l
	}
}

func (d *Doc) Height() int {
	return d.height
}

func (d *Doc) textHeight() int {
	if d.auto {
		return d.height - 9
	}
	return d.height - 1
}

func (d *Doc) statusBarY() int {
	return d.height
}

func (d *Doc) predictionsY() int {
	return d.height - 8
}

func (d *Doc) predictionsHeight() int {
	if d.auto {
		return 8
	}
	return 0
}

func (d *Doc) Width() int {
	return d.width
}

func (d *Doc) textWidth() int {
	return d.width - 1
}
