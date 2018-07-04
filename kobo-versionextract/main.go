package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

var version = "dev"

func main() {
	help := pflag.BoolP("help", "h", false, "show this help text")
	pflag.Parse()

	if *help || pflag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: kobo-versionextract [OPTIONS] PATH_TO_FW\n")
		fmt.Fprintf(os.Stderr, "\nVersion: %s\n\nOptions:\n", version)
		pflag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nPATH_TO_FW is either the path to a KoboRoot.tgz, libnickel.so, or a kobo update zip.\n\nNote that kobo-versionextract only works with firmware 4.7.10413 and later.\n")
		os.Exit(1)
	}

	fwpath := pflag.Arg(0)
	if _, err := os.Stat(fwpath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Specified path does not exist.\n")
		os.Exit(1)
	}

	var libnickel io.Reader
	var revinfo io.Reader
	var tgz io.Reader
	var date *time.Time
	if strings.Contains(fwpath, "libnickel.so") {
		var err error
		libnickel, err = os.Open(fwpath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Could not open file: %v\n", err)
			os.Exit(1)
		}
		defer libnickel.(*os.File).Close()
	} else if strings.HasSuffix(fwpath, ".zip") {
		zr, err := zip.OpenReader(fwpath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Could not open zip file: %v\n", err)
			os.Exit(1)
		}
		defer zr.Close()

		for _, f := range zr.File {
			if f.Name == "KoboRoot.tgz" {
				tgz, err = f.Open()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: Could not open KoboRoot.tgz: %v\n", err)
					os.Exit(1)
				}
				defer tgz.(io.ReadCloser).Close()
			}
		}
	} else if strings.HasSuffix(fwpath, ".tgz") {
		var err error
		tgz, err = os.Open(fwpath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Could not open file: %v\n", err)
			os.Exit(1)
		}
		defer tgz.(*os.File).Close()
	} else {
		fmt.Fprintf(os.Stderr, "Error: Could not detect fw type (doesn't have any of the following suffixes: .so*, .zip, .tgz).\n")
		os.Exit(1)
	}

	if tgz != nil {
		zr, err := gzip.NewReader(tgz)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Could not open KoboRoot.tgz as gzip: %v\n", err)
			os.Exit(1)
		}
		defer zr.Close()

		tr := tar.NewReader(zr)
		for {
			h, err := tr.Next()
			if err == io.EOF {
				err = nil
				break
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Could not read KoboRoot.tgz: %v\n", err)
				os.Exit(1)
			}

			if strings.HasSuffix(h.Name, "usr/local/Kobo/libnickel.so.1.0.0") {
				buf, err := ioutil.ReadAll(tr)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: Could not read libnickel from KoboRoot.tgz: %v\n", err)
					os.Exit(1)
				}
				libnickel = bytes.NewReader(buf)
			} else if strings.HasSuffix(h.Name, "usr/local/Kobo/revinfo") {
				ct := h.ModTime
				date = &ct
				buf, err := ioutil.ReadAll(tr)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: Could not read revinfo from KoboRoot.tgz: %v\n", err)
					os.Exit(1)
				}
				revinfo = bytes.NewReader(buf)
			}
		}
	}

	if revinfo != nil {
		buf, err := ioutil.ReadAll(revinfo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Could not read revinfo: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Revision: %s\n", strings.TrimSpace(string(buf)))
	}

	if libnickel == nil {
		fmt.Fprintf(os.Stderr, "Error: Could not find libnickel\n")
		os.Exit(1)
	}

	buf, err := ioutil.ReadAll(libnickel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not read libnickel: %v\n", err)
		os.Exit(1)
	}

	i := bytes.Index(buf, []byte("Kobo Touch %2/%3"))
	if i < 0 {
		fmt.Fprintf(os.Stderr, "Error: Could not find user agent string in libnickel. Please report this as a bug.\n")
		os.Exit(1)
	}

	matches := regexp.MustCompile(`[0-9]+\.[0-9]+\.[0-9]+`).FindAll(buf[i:i+200], -1)
	if len(matches) > 1 {
		fmt.Printf("Warning: more than 1 regexp match\n")
	} else if len(matches) < 1 {
		fmt.Printf("Error: no regexp match for version between user agent string and 200 chars after\n")
		os.Exit(1)
	}
	fmt.Printf(" Version: %s\n", matches[0])

	if date != nil {
		fmt.Printf("    Date: %s\n", (*date).Format("January 2006"))
	}

	if runtime.GOOS == "windows" {
		time.Sleep(time.Second * 4)
	}
}
