// Package bsearch implements efficient random-access lookups over delimited
// files, such as Unix newline-delimited records.
package bsearch

import (
	"fmt"
	"io"
)

// Reader is a file.
type Reader interface {
	ReadAt(b []byte, off int64) (n int, err error)
}

// Record is file portion.
type Record struct {
	// The absolute byte offset of the record inside the file.
	Start, End int64
	// The record payload, including the delimiter.
	Data []byte
	// An error, if the record could not be read.
	Err error
}

func (r Record) String() string {
	return fmt.Sprintf("Record{Extent=[%v:%v] Data=%q}", r.Start, r.End, string(r.Data))
}

// Config configures Read operations.
type Config struct {
	// The amount of data to read for each call to ReadAt().
	ChunkSize int
	// The record delimiter, typically \n.
	Delimiter byte

	// Returns whether lhs is less than rhs.
	Less func(lhs, rhs []byte) bool
}

// Read reads a record at the given position.
func Read(c Config, r Reader, pos int64) Record {
	var (
		res      Record
		first    []byte = make([]byte, c.ChunkSize)
		start           = pos - int64(c.ChunkSize)/2
		posInCur        = pos - start
		fail            = func(err error) Record {
			res.Err = err
			return res
		}
	)
	if start < 0 {
		posInCur += start
		start = 0
	}

	// Try and fetch an entire record in a single read system call.
	n, err := r.ReadAt(first, start)
	if err != nil && err != io.EOF {
		return fail(err)
	}
	first = first[:n]
	if len(first) == 0 {
		return fail(io.EOF)
	}

	// Find the parts of the record before and after pos, then join them together.
	sChunk, err := startChunk(c, r, first[:posInCur], pos)
	if err != nil {
		return fail(err)
	}
	eChunk, err := endChunk(c, r, first[posInCur:], pos)

	res.Data = append(sChunk, eChunk...)
	res.Err = err
	res.Start = pos - int64(len(sChunk))
	res.End = pos + int64(len(eChunk))
	return res
}

// endChunk reads from r until it can return a slice containing c.Delimiter.
// Returns io.EOF if it reaches the end of the file without finding the delimiter.
func endChunk(c Config, r Reader, buf []byte, start int64) ([]byte, error) {
	var (
		cur = buf
		off int
		err error
		n   int
	)
	for {
		for i, b := range cur {
			// If io.EOF on our last read but we find the delimiter, return nil error.
			if b == c.Delimiter {
				return buf[:off+i+1], nil
			}
		}
		if err != nil {
			return nil, err
		}
		off += len(cur)
		cur = make([]byte, c.ChunkSize)
		n, err = r.ReadAt(cur, start+int64(off))
		cur = cur[:n]
		buf = append(buf, cur...)
	}
}

// startChunk reads backwards from r until it can return a slice containing c.Delimiter, or it reaches the start of the file.
func startChunk(c Config, r Reader, buf []byte, end int64) ([]byte, error) {
	cur := buf
	start := end - int64(len(buf))
	for {
		for i := len(cur) - 1; i >= 0; i-- {
			b := cur[i]
			if b == c.Delimiter {
				return buf[i+1:], nil
			}
		}
		if len(cur) == 0 {
			return buf, nil
		}
		size := int64(c.ChunkSize)
		start -= size
		if start < 0 {
			size += start
			start = 0
		}
		cur = make([]byte, size)
		n, err := r.ReadAt(cur, start)
		if err != nil {
			return nil, err
		}
		cur = cur[:n]
		buf = append(cur, buf...)
	}
}

// LowerBound returns the first element not less than value.
func LowerBound(c Config, r Reader, end int64, value []byte) Record {
	if c.Less == nil {
		panic("c.Less must be set.")
	}
	var (
		last  Record
		start int64 = 0
	)
	for {
		var (
			width = (end - start)
			pos   = start + width/2
			res   = Read(c, r, pos)
		)
		if res.Err != nil {
			return res
		}
		if c.Less(res.Data, value) {
			start = res.End
			last = res
		} else {
			end = res.Start
		}

		if start == end {
			break
		}
	}
	return last
}

// UpperBound returns the first element greater than value.
func UpperBound(c Config, r Reader, end int64, value []byte) Record {
	if c.Less == nil {
		panic("c.Less must be set.")
	}
	var (
		first Record
		start int64 = 0
	)
	for {
		var (
			width = (end - start)
			pos   = start + width/2
			res   = Read(c, r, pos)
		)
		if res.Err != nil {
			return res
		}
		if c.Less(value, res.Data) {
			first = res
			end = res.Start
		} else {
			start = res.End
		}

		if start == end {
			break
		}
	}
	return first
}
