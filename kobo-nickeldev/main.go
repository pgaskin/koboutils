package main

import (
	"debug/elf"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/pgaskin/koboutils/v2/internal"
	"github.com/spf13/pflag"
)

func main() {
	help := pflag.BoolP("help", "h", false, "show this help text")
	pflag.Parse()

	if *help || pflag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "usage: kobo-nickeldev [options] libnickel_path...\n")
		fmt.Fprintf(os.Stderr, "\nversion: %s\n\noptions:\n", internal.VersionName())
		pflag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\ncompatibility: works with 20400+, tested up to 175773\n")
		os.Exit(2)
	}

	f, err := os.Open(pflag.Arg(0))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	ds, err := scan(f)
	if err != nil {
		panic(err)
	}

	fmt.Printf("| %-36s | %-20s | %-20s | %-26s |\n", "ID", "Codename", "Family", "Name")
	fmt.Printf("| %-36s | %-20s | %-20s | %-26s |\n", "---", "---", "---", "---")
	for _, d := range ds {
		fmt.Printf("| %-36s | %-20s | %-20s | %-26s |\n", d.ID, d.Codename, d.Family, d.Name)
	}
}

type DeviceInfo struct {
	Codename string
	Family   string
	ID       string
	Name     string
}

// scan reads the device info array from the provided libnickel elf.
func scan(f io.ReaderAt) ([]DeviceInfo, error) {
	ef, err := elf.NewFile(f)
	if err != nil {
		return nil, err
	}
	defer ef.Close()

	efm := internal.ELFMem{Elf: ef}

	// struct used by nickel for basic device info
	type deviceInfoStruct struct {
		Codename uint32 // char*
		Family   uint32 // char*
		ID       uint32 // char*
		Name     uint32 // char*
	}
	readDeviceInfo := func(vaddr uint32) (DeviceInfo, bool, error) {
		var (
			err error
			out DeviceInfo
			obj deviceInfoStruct
		)
		if err = efm.ReadStructAt(uint64(vaddr), binary.LittleEndian, &obj); err != nil {
			return out, false, err // error
		}
		if obj.Codename == 0 || obj.Family == 0 || obj.ID == 0 || obj.Name == 0 {
			return out, false, nil
		}
		if out.Codename, err = efm.ReadCString(uint64(obj.Codename), 128); err != nil {
			return out, false, err // error
		}
		if out.Family, err = efm.ReadCString(uint64(obj.Family), 128); err != nil {
			return out, false, err // error
		}
		if out.ID, err = efm.ReadCString(uint64(obj.ID), 128); err != nil {
			return out, false, err // error
		} else if !strings.HasPrefix(out.ID, "00000000-0000-0000-0000-000000000") {
			return out, false, fmt.Errorf("invalid id")
		}
		if out.Name, err = efm.ReadCString(uint64(obj.Name), 128); err != nil {
			return out, false, err // error
		}
		return out, true, nil
	}

	// find a known device
	var (
		idVaddrs        []uint64
		trilogyVaddrs   []uint64
		koboTouchVaddrs []uint64
	)
	if err := efm.EachIndex([]byte("00000000-0000-0000-0000-000000000310\x00"), 0, func(vaddr uint64) bool {
		idVaddrs = append(idVaddrs, vaddr)
		return true
	}); err != nil {
		return nil, fmt.Errorf("scan for kobo touch device id string: %w", err)
	} else if len(idVaddrs) == 0 {
		return nil, fmt.Errorf("scan for kobo touch device id string: not found")
	}
	if err := efm.EachIndex([]byte("trilogy\x00"), 0, func(vaddr uint64) bool {
		trilogyVaddrs = append(trilogyVaddrs, vaddr)
		return true
	}); err != nil {
		return nil, fmt.Errorf("scan for trilogy string: %w", err)
	} else if len(trilogyVaddrs) == 0 {
		return nil, fmt.Errorf("scan for trilogy string: not found")
	}
	if err := efm.EachIndex([]byte("Kobo Touch\x00"), 0, func(vaddr uint64) bool {
		koboTouchVaddrs = append(koboTouchVaddrs, vaddr)
		return true
	}); err != nil {
		return nil, fmt.Errorf("scan for kobo touch string: %w", err)
	} else if len(koboTouchVaddrs) == 0 {
		return nil, fmt.Errorf("scan for kobo touch string: not found")
	}

	// find the struct for it
	var koboTouchInfoStructVaddr uint64 = math.MaxUint64
koboTouchInfoStructVaddr:
	for _, idVaddr := range idVaddrs {
		for _, trilogyVaddr := range trilogyVaddrs {
			for _, koboTouchVaddr := range koboTouchVaddrs {
				obj := deviceInfoStruct{
					Codename: uint32(trilogyVaddr),
					Family:   uint32(trilogyVaddr),
					ID:       uint32(idVaddr),
					Name:     uint32(koboTouchVaddr),
				}
				idx, err := efm.IndexStruct(binary.LittleEndian, obj, 0)
				if err != nil {
					return nil, fmt.Errorf("scan for kobo touch device info struct: %w", err)
				}
				if idx != math.MaxUint64 {
					koboTouchInfoStructVaddr = idx
					break koboTouchInfoStructVaddr
				}
			}
		}
	}
	if koboTouchInfoStructVaddr == math.MaxUint64 {
		return nil, fmt.Errorf("scan for kobo touch device info struct: not found")
	}

	// extend backwards until we don't have a valid one
	deviceInfoBaseVaddr := koboTouchInfoStructVaddr
	for {
		if _, ok, err := readDeviceInfo(uint32(deviceInfoBaseVaddr) - uint32(binary.Size(deviceInfoStruct{}))); !ok || err != nil {
			break
		}
		deviceInfoBaseVaddr -= uint64(binary.Size(deviceInfoStruct{}))
	}

	// read forwards until the null terminator
	var ds []DeviceInfo
	for vaddr := deviceInfoBaseVaddr; ; vaddr += uint64(binary.Size(deviceInfoStruct{})) {
		d, ok, err := readDeviceInfo(uint32(vaddr))
		if err != nil {
			return nil, fmt.Errorf("read kobo touch device info array: element at 0x%X: %w", vaddr, err)
		}
		if !ok {
			break
		}
		ds = append(ds, d)
	}
	return ds, nil
}
