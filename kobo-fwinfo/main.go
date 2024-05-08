package main

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"os"

	"github.com/pgaskin/koboutils/v2/internal"
	"github.com/spf13/pflag"
)

// TODO: extract other files into a package

func main() {
	help := pflag.BoolP("help", "h", false, "show this help text")
	tempDir := pflag.StringP("temp-dir", "t", "", "override the temp dir for extracting large firmware files to")
	pflag.Parse()

	if *help || pflag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "usage: kobo-fwinfo [options] firmware_path...\n")
		fmt.Fprintf(os.Stderr, "\nversion: %s\n\noptions:\n", internal.VersionName())
		pflag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nfirmware_path is the path to a kobo update package, optionally unpacked.\n")
		os.Exit(2)
	}

	var errs int
	for _, fw := range pflag.Args() {
		var ok bool
		p, err := func() (Package, error) {
			var p Package

			fi, err := os.Stat(fw)
			if err != nil {
				return p, err
			}

			var fsys fs.FS
			if fi.IsDir() {
				fsys = os.DirFS(fw)
			} else {
				z, err := zip.OpenReader(fw)
				if err != nil {
					return p, err
				}
				defer z.Close()
				fsys = z
			}

			ok = true
			return Parse(fsys, ReadTempDirAuto(*tempDir))
		}()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: error: %v\n", fw, err)
			errs++
		}
		if ok {
			fmt.Printf("%s: %s\n", fw, p)
		}
	}
	if errs != 0 {
		os.Exit(1)
	}
}
