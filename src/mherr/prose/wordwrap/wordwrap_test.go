package wordwrap

import (
	"reflect"
	"testing"
)

func TestFold(t *testing.T) {
	tests := []struct {
		desc  string
		input string
		want  []string
	}{
		{
			"empty",
			"",
			nil,
		},
		{
			"single word",
			"hello",
			[]string{"hello"},
		},
		{
			"splitting a long line",
			"the quick brown fox jumped over the lazy dog",
			[]string{"the quick", "brown fox", "jumped", "over the", "lazy dog"},
		},
		{
			"handling newlines",
			"the fox.\nthe quick.",
			[]string{"the fox.", "", "the quick."},
		},
		{
			"preformatted blocks followed by paragraphs",
			" * foo\n * bar\nla rutrum",
			[]string{" * foo", " * bar", "", "la rutrum"},
		},
		{
			"preformatted blocks followed by paragraphs (2)",
			" * foo\n * bar\nla rutrum more text more",
			[]string{" * foo", " * bar", "", "la rutrum", "more text", "more"},
		},
		{
			"leading and duplicate newlines are ignored",
			"\n\nthe fox.\n\n\n\nthe quick.",
			[]string{"the fox.", "", "the quick."},
		},
		{
			"embedded spaces are preserved",
			"  hello\n  world\n",
			[]string{"  hello", "  world"},
		},
		{
			"edge case for line ending",
			"0123456789\n",
			[]string{"0123456789"},
		},
		{
			"edge case for line ending (2)",
			"012345678\n01234",
			[]string{"012345678", "", "01234"},
		},
	}

	for _, c := range tests {
		/*
			if c.desc != "handling newlines" {
				continue
			}
		*/
		got := Fold(c.input, 10)
		if !reflect.DeepEqual(got, c.want) {
			t.Fatalf("test(%v): got:\n%#v\nwant:\n%#v", c.desc, got, c.want)
		}
	}
}

func TestUnfold(t *testing.T) {
	tests := []struct {
		desc  string
		input []string
		want  string
	}{
		{
			"empty",
			nil,
			"",
		},
		{
			"single line",
			[]string{"hello"},
			"hello\n",
		},
		{
			"paragraphs",
			[]string{"Hello", "world.", "", "The end."},
			"Hello world.\n\nThe end.\n",
		},
		{
			"preformatted blocks",
			[]string{"A few notes:", "  Hello", "  Things.", "", "The end."},
			"A few notes:\n\n  Hello\n  Things.\n\nThe end.\n",
		},
	}

	for _, c := range tests {
		/*
			if c.desc != "handling newlines" {
				continue
			}
		*/
		got := string(Unfold(c.input))
		if got != c.want {
			t.Fatalf("test(%v): got:\n%#v\nwant:\n%#v", c.desc, got, c.want)
		}
	}
}
