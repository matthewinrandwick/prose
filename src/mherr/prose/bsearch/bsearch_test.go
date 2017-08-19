package bsearch

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestRead(t *testing.T) {
	tests := []struct {
		desc string
		data string
		pos  int64
		want Record
	}{
		{
			desc: "empty file has no trailing delimiter, bad",
			want: Record{Err: io.EOF},
		},
		{
			desc: "minimal file with just a delimiter",
			data: ".",
			pos:  0,
			want: Record{Start: 0, End: 1, Data: []byte(".")},
		},
		{
			desc: "two records",
			data: "..",
			pos:  1,
			want: Record{Start: 1, End: 2, Data: []byte(".")},
		},
		{
			desc: "three records",
			data: "...",
			pos:  2,
			want: Record{Start: 2, End: 3, Data: []byte(".")},
		},
		{
			desc: "four records",
			data: "....",
			pos:  3,
			want: Record{Start: 3, End: 4, Data: []byte(".")},
		},
		{
			desc: "a long record, reading a long tail",
			data: "a.01234567890.abcdef",
			pos:  3,
			want: Record{Start: 2, End: 14, Data: []byte("01234567890.")},
		},
		{
			desc: "a long record, reading a long head",
			data: "a.01234567890.abcdef",
			pos:  11,
			want: Record{Start: 2, End: 14, Data: []byte("01234567890.")},
		},
		{
			desc: "a long record, reading both a long head and tail",
			data: "abc.012345678901234567890.abc.",
			pos:  10,
			want: Record{Start: 4, End: 26, Data: []byte("012345678901234567890.")},
		},
	}
	cfg := Config{
		ChunkSize: 4,
		Delimiter: '.',
	}

	for _, c := range tests {
		r := bytes.NewReader([]byte(c.data))
		data := Read(cfg, r, c.pos)
		if !reflect.DeepEqual(data, c.want) {
			t.Fatalf("test(%v): bad result: %v", c.desc, data)
		}
	}
}

func TestLowerBound(t *testing.T) {
	tests := []struct {
		desc string
		data string
		req  string
		want Record
	}{
		{
			desc: "empty input",
			want: Record{Err: io.EOF},
		},
		{
			desc: "longer example",
			data: "01.02.03.04.05.06.07.08.09.10.",
			req:  "02x.",
			want: Record{Start: 3, End: 6, Data: []byte("02.")},
		},
		{
			desc: "finds the first _less_ than the request",
			data: "01.02.03.04.05.06.07.08.09.10.",
			req:  "04.",
			want: Record{Start: 6, End: 9, Data: []byte("03.")},
		},
		{
			desc: "no result",
			data: "01.02.03.04.05.06.07.08.09.10.",
			req:  "01.",
			want: Record{Start: 0, End: 0, Data: nil},
		},
		{
			desc: "find last",
			data: "01.02.03.04.05.06.07.08.09.10.",
			req:  "10x.",
			want: Record{Start: 27, End: 30, Data: []byte("10.")},
		},
	}
	cfg := Config{
		ChunkSize: 4,
		Delimiter: '.',
		Less: func(l, r []byte) bool {
			return string(l) < string(r)
		},
	}

	for _, c := range tests {
		r := bytes.NewReader([]byte(c.data))
		data := LowerBound(cfg, r, int64(len(c.data)), []byte(c.req))
		if !reflect.DeepEqual(data, c.want) {
			t.Fatalf("test(%v): bad result: %v", c.desc, data)
		}
	}
}

func TestUpperBound(t *testing.T) {
	tests := []struct {
		desc string
		data string
		req  string
		want Record
	}{
		{
			desc: "empty input",
			want: Record{Err: io.EOF},
		},
		{
			desc: "longer example",
			data: "01.02.03.04.05.06.07.08.09.10.",
			req:  "02x.",
			want: Record{Start: 6, End: 9, Data: []byte("03.")},
		},
		{
			desc: "finds the first greater than the request",
			data: "01.02.03.04.05.06.07.08.09.10.",
			req:  "04.",
			want: Record{Start: 12, End: 15, Data: []byte("05.")},
		},
		{
			desc: "match first entry",
			data: "01.02.03.04.05.06.07.08.09.10.",
			req:  "",
			want: Record{Start: 0, End: 3, Data: []byte("01.")},
		},
		{
			desc: "match last entry (fails, none greater)",
			data: "01.02.03.04.05.06.07.08.09.10.",
			req:  "10.",
			want: Record{Start: 0, End: 0, Data: nil},
		},
	}
	cfg := Config{
		ChunkSize: 4,
		Delimiter: '.',
		Less: func(l, r []byte) bool {
			return string(l) < string(r)
		},
	}

	for _, c := range tests {
		r := bytes.NewReader([]byte(c.data))
		data := UpperBound(cfg, r, int64(len(c.data)), []byte(c.req))
		if !reflect.DeepEqual(data, c.want) {
			t.Fatalf("test(%v): bad result: %v", c.desc, data)
		}
	}
}
