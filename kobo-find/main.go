package main

import (
	"fmt"
	"os"

	"github.com/geek1011/koboutils/kobo"
	"github.com/spf13/pflag"
)

func main() {
	first := pflag.BoolP("first", "f", false, "only show the first kobo detected")
	help := pflag.BoolP("help", "h", false, "show this help text")
	pflag.Parse()

	if *help || pflag.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "Usage: kobo-find [OPTIONS]\n\nOptions:\n")
		pflag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nkobo-find requires the findmnt command on linux.\n")
		os.Exit(1)
	}

	kobos, err := kobo.Find()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(kobos) < 1 {
		os.Exit(1)
	}

	for _, kobo := range kobos {
		fmt.Println(kobo)
		if *first {
			break
		}
	}
}
