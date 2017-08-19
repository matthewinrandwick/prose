// Package conio handles console I/O.
package conio

import (
	"bytes"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	Stdout = 0

	ClearScreen = "2J"
	Home        = "H"
	ClearLine   = "2K"

	tcGets = 0x5401 // syscall.TCGETS
	tcSets = 0x5402 // syscall.TCSETS

	CodeEsc = 27
)

// State contains the state of a terminal.
type State struct {
	syscall.Termios
}

// Raw sets the terminal to raw mode.
func Raw() (*State, error) {
	var oldState State
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, Stdout, tcGets, uintptr(unsafe.Pointer(&oldState.Termios)), 0, 0, 0); err != 0 {
		return nil, err
	}

	newState := oldState.Termios
	// This attempts to replicate the behaviour documented for cfmakeraw in
	// the Termios(3) manpage.
	newState.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP | syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	newState.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	newState.Cflag &^= syscall.CSIZE | syscall.PARENB
	newState.Cflag |= syscall.CS8
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, Stdout, tcSets, uintptr(unsafe.Pointer(&newState)), 0, 0, 0)
	return &oldState, cast(err)
}

// Restore restores the terminal connected to the given file descriptor to a
// previous state.
func Restore(state *State) error {
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, Stdout, tcSets, uintptr(unsafe.Pointer(&state.Termios)), 0, 0, 0)
	return cast(err)
}

func cast(err syscall.Errno) error {
	if err == 0 {
		return nil
	}
	return err
}

// Size returns the dimensions of the given terminal.
func Size() (width, height int, err error) {
	var dimensions [4]uint16
	_, _, e := syscall.Syscall6(syscall.SYS_IOCTL, Stdout, uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&dimensions)), 0, 0, 0)
	return int(dimensions[1]), int(dimensions[0]), cast(e)
}

// char returns a single character.
func char() byte {
	bs := make([]byte, 1)
	n, err := tty.Read(bs)
	if err != nil {
		panic(err)
	}
	if n != 1 {
		panic("could not read from tty")
	}
	return bs[0]
}

// Returns a single character, or an ANSI escape sequence from the tty.
func Seq() string {
	var buf bytes.Buffer
	c := char()
	buf.WriteByte(c)
	if c != CodeEsc {
		return buf.String()
	}

	c = char()
	buf.WriteByte(c)
	switch c {
	case '[':
	case 'O':
	case CodeEsc:
		c = char()
		buf.WriteByte(c)
	default:
		return buf.String()
	}

	for {
		c = char()
		buf.WriteByte(c)
		// ANSI escape sequences end with a character in this range.
		if c >= 64 && c <= 126 {
			return buf.String()
		}
	}
}

var tty *os.File

func init() {
	var err error
	tty, err = os.Open("/dev/tty")
	if err != nil {
		panic(err)
	}
}

// Escape writes out an escape sequence.
func Escape(t string) {
	fmt.Printf("\x1b[%v", t)
}

// Escapef writes out an escape sequence.
func Escapef(pat string, args ...interface{}) {
	Escape(fmt.Sprintf(pat, args...))
}

// Out writes the output to the terminal.
func Out(t string) {
	fmt.Print(t)
}

// Outf writes out text to the terminal.
func Outf(pat string, args ...interface{}) {
	fmt.Printf(pat, args...)
}

// Pos positions the cursor at the given y, x location..
func Pos(y, x int) {
	Escapef("%v;%vH", y, x)
}
