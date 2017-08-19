package main

import (
	"errors"
	"flag"
	"fmt"
	"mherr/prose/conio"
	"mherr/prose/ngram"
	"mherr/prose/view"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var errExit = errors.New("exit requested")

func usage() {
	flag.Usage()
	os.Exit(2)
}

func main() {
	ngram.ResourcePath = filepath.Dir(os.Args[0])

	flag.Usage = func() { fmt.Print("usage: prose [filename]\n") }
	flag.Parse()
	filename := flag.Arg(0)
	if filename == "" {
		usage()
	}

	t, err := conio.Raw()
	if err != nil {
		panic(err)
	}
	fail := func(err error) {
		conio.Restore(t)
		conio.Escape(conio.ClearScreen)
		conio.Escape(conio.Home)
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := conio.Restore(t); err != nil {
			panic(err)
		}
	}()

	conio.Escape(conio.ClearScreen)
	conio.Escape(conio.Home)

	seq := pollTerminal()
	winChanged := make(chan os.Signal)
	signal.Notify(winChanged, syscall.SIGWINCH)

	d, err := view.New(filename)
	if err != nil {
		fail(err)
	}

	d.Redraw()

out:
	for {
		var s string
		select {
		case <-winChanged:
			d.WindowChanged()

		case s = <-seq:
			err := handleKeypress(s, d, seq)
			if err == errExit {
				break out
			}
			if err != nil {
				fail(err)
			}
		}
	}

	conio.Escape(conio.ClearScreen)
	conio.Escape(conio.Home)
}

func handleKeypress(s string, d *view.Doc, seq chan string) error {
	switch {
	case s == "\x01": // Control-A
		d.Auto(true)
	case s == "\x0f": // Control-O
		d.Auto(false)
	case s == "\x1b[A":
		d.Move(-1, 0)
	case s == "\x1b[B":
		d.Move(1, 0)
	case s == "\x1b[C":
		d.Move(0, 1)
	case s == "\x1b[D":
		d.Move(0, -1)
	case s == "\x03": // Control-C
		fallthrough
	case s == "\x04": // Control-D
		if d.Dirty() {
			d.WriteStatus("Changes not saved, exit anyway? (y/N)")
			s := <-seq
			if s == "y" {
				return errExit
			}
			d.Redraw()
		} else {
			return errExit
		}
	case s == "\x1b[1~": // Home
		d.Move(0, -900)
	case s == "\x1b[4~": // End
		d.Move(0, 900)
	case s == "\x1b[5~": // Page Up
		d.Move(-d.Height()*3/2, 0)
	case s == "\x1b[6~": // Page Down
		d.Move(d.Height()*3/2, 0)
	case s == "\t":
		return d.Edit(s[0])
	case len(s) == 1 && s[0] >= 32 && s[0] < 127:
		return d.Edit(s[0])
	case s == "\r": // Enter
		d.Enter()
	case s == "\x13": // Save
		return d.Save()
	case s == "\b":
		d.CtlBackspace()
	case s == "\x7f":
		d.Backspace()
	case s == "\x1b[3~": // Delete
		d.Delete()
	default:
		conio.Escape(conio.Home)
		conio.Escape(conio.ClearLine)
		fmt.Printf("seq=%q\n", s)
	}
	return nil
}

func pollTerminal() chan string {
	ch := make(chan string)
	go func() {
		for {
			ch <- conio.Seq()
		}
	}()
	return ch
}
