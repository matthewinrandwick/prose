package ngram

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestPredictions(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"he", "re"},
		{"th", "e"},
		{"other", " relevant"},
	}

	for _, c := range tests {
		got, err := Predictions(c.input)
		if err != nil {
			t.Fatal(err)
		}
		var top Match
		if len(got) > 0 {
			top = got[0]
		}
		if top.Text != c.want {
			t.Fatalf("for input %q, got:%v\nwant: %v", c.input, got, c.want)
		}
	}
}

const text = `1	a
`

func TestFind(t *testing.T) {
	tests := []struct {
		desc    string
		file    string
		sought  string
		want    Matches
		wantErr bool
	}{
		{
			desc:   "missing search returns empty slice",
			file:   "",
			sought: "missing",
			want:   nil,
		},
		{
			desc: "simple example",
			file: "" +
				"a\t0\n" +
				"acacia\t25\n" +
				"acorn\t20\n" +
				"acted\t30\n" +
				"apiary\t30\n" +
				"beaver\t5\n",
			sought: "ac",
			want: Matches{
				{"acted", 30, 1},
				{"acacia", 25, 1},
				{"acorn", 20, 1},
			},
		},
	}
	for _, c := range tests {
		content := []byte(c.file)
		tmp, err := ioutil.TempFile("", "Find")
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			tmp.Close()
			os.Remove(tmp.Name())
		}()
		if _, err := tmp.Write([]byte(content)); err != nil {
			t.Fatalf("test(%v): %v", c.desc, err)
		}
		got, err := Find(tmp.Name(), c.sought, 1)
		if (err != nil) != c.wantErr {
			t.Fatalf("test(%v): bad err: %v wantErr: %v", c.desc, err, c.wantErr)
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Fatalf("test(%v): bad matches got: %v", c.desc, got)
		}
	}
}

func init() {
	ResourcePath = "../bin"
}
