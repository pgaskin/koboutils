package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/geek1011/koboutils/kobo"
	"github.com/spf13/pflag"
)

var jsono = false

func main() {
	json := pflag.BoolP("json", "j", false, "output as json")
	help := pflag.BoolP("help", "h", false, "show this help text")
	pflag.Parse()

	if *help || pflag.NArg() > 1 {
		fmt.Fprintf(os.Stderr, "Usage: kobo-info [OPTIONS] [KOBO_PATH]\n\nOptions:\n")
		pflag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nIf KOBO_PATH is not specified, kobo-info will attempt to look for a kobo device.\n")
		os.Exit(1)
	}

	jsono = *json

	var kpath string
	if pflag.NArg() == 1 {
		kpath = pflag.Arg(0)
	} else {
		kobos, err := kobo.Find()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not look for a kobo: %v\n", err)
			os.Exit(1)
		} else if len(kobos) < 1 {
			fmt.Fprintf(os.Stderr, "Error: could not find a kobo\n")
			os.Exit(1)
		}
		kpath = kobos[0]
	}

	if !kobo.IsKobo(kpath) {
		fmt.Fprintf(os.Stderr, "Error: not a valid kobo: %s\n", kpath)
		os.Exit(1)
	}

	serial, version, id, err := kobo.ParseKoboVersion(kpath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not parse kobo version: %v\n", err)
		os.Exit(1)
	}

	if device, ok := kobo.DeviceByID(id); ok {
		printkv("Device", device.Name)
		printkv("Device ID", id)
		printkv("Hardware", device.Hardware)
	} else {
		printkv("Device", "unknown")
		printkv("Device ID", id)
	}

	println()
	printkv("Serial", serial)
	println()
	printkv("Current FW", version)

	if affiliate, err := kobo.ParseKoboAffiliate(kpath); err == nil {
		printkv("Affiliate", affiliate)
	} else {
		printkv("Affiliate", "unknown")
	}

	if jsono {
		fmt.Print("\n}\n")
	}

	if runtime.GOOS == "windows" {
		time.Sleep(time.Second * 5)
	}
}

var jsons = true

func printkv(key, value string) {
	if jsono {
		if jsons {
			fmt.Print("{\n")
			jsons = false
		} else {
			fmt.Print(",\n")
		}
		fmt.Printf(`    "%s": "%s"`, strings.Replace(strings.ToLower(key), " ", "_", -1), value)
	} else {
		fmt.Printf("%10s: %s\n", key, value)
	}
}

func println() {
	if !jsono {
		fmt.Print("\n")
	}
}
