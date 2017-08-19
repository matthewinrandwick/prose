package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
)

func usage() {
	flag.Usage()
	os.Exit(2)
}

func read(f *os.File) error {
	d := xml.NewDecoder(f)

	for {
		t, err := d.Token()
		if err != nil {
			return err
		}
		switch t := t.(type) {
		case xml.StartElement:
			n := t.Name.Local
			switch n {
			case "s":
				fmt.Printf("\n")
			case "c":
				t, err := d.Token()
				if err != nil {
					return err
				}
				d, ok := t.(xml.CharData)
				if ok {
					fmt.Print(string(d))
				}
			case "w":
				t, err := d.Token()
				if err != nil {
					return err
				}
				d, ok := t.(xml.CharData)
				if ok {
					fmt.Print(string(d))
				}
			}
		}
	}
	return nil
}

func main() {
	flag.Usage = func() { fmt.Print("usage: bnc-to-text [filename.xml...]\n") }
	flag.Parse()

	filenames := flag.Args()
	if len(filenames) == 0 {
		usage()
	}
	for _, name := range filenames {
		f, err := os.Open(name)
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}

		if err := read(f); err != nil {
			if err == io.EOF {
				fmt.Println()
				continue
			}
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		f.Close()
	}
}
