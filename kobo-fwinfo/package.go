package main

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/klauspost/compress/gzip"
)

// Package contains metadata for a firmware update package.
type Package struct {
	Format   PackageFormat
	Version  Version
	Branch   string
	Revision string
	Date     time.Time
}

// PackageFormat is a firmware update package layout.
type PackageFormat int

const (
	PackageFormatUnknown = iota

	// PackageFormatKobo is an update distributed as a Kobo.tgz.
	//
	// The Kobo.tgz is placed in the .kobo folder.
	//
	// This is not currently used for official firmware updates.
	//
	// On reboot:
	//  - If "pickel can-upgrade" doesn't return 0 (e.g., due to low battery), the upgrade is ignored and boot continues normally.
	//  - The contents are extracted into /usr/local/Kobo.
	//  - The update files are removed.
	//  - Boot continues normally.
	PackageFormatKobo

	// PackageFormatKoboRoot is an update distributed as a KoboRoot.tgz.
	//
	// The KoboRoot.tgz is placed in the .kobo folder. Updates to the firmware
	// blobs are placed in product-specific subdirs of the .kobo/upgrade folder.
	// MD5 checksums are optionally placed in manifest.md5sum.
	//
	// This format is used for all firmware updates until firmware v5.
	//
	// On reboot:
	//  - If "pickel can-upgrade" doesn't return 0 (e.g., due to low battery), the upgrade is ignored and boot continues normally.
	//  - The boot partition is set to recoveryfs (to trigger a factory reset on failure).
	//  - The update animation is started.
	//  - If manifest.md5sum exists, it is checked using `md5sum -c manifest.md5sum`, then removed.
	//  - Firmware blobs are updated from the product-specific subdirectory if it exists, then removed.
	//  - KoboRoot.tgz is extracted into /, then removed.
	//  - The update animation is stopped.
	//  - The boot partition is set to rootfs.
	//  - The device is rebooted.
	PackageFormatKoboRoot

	// PackageFormatGeneric is an update distributed as an update.tar.
	//
	// The update.tar is placed in the .kobo folder. This format is supported by
	// v5 and later, and is mostly freeform.
	//
	// The tar may be compressed with any format supported by the tar command.
	//
	// It appears to depend on the new recovery partition used on devices
	// shipped with v5.
	//
	// On reboot:
	//  - If "pickel can-upgrade" doesn't return 0 (e.g., due to low battery), the upgrade is ignored and boot continues normally.
	//  - The driver.sh script is extracted into /tmp/update.
	//  - /tmp/update/driver.sh /path/to/update.tar stage1 PRODUCT
	//  - The boot partition is set to recoveryfs for the next boot, and the device is rebooted.
	//  - Partitions are mounted normally.
	//  - If the update.tar file doesn't exist, a hard factory reset is triggered.
	//  - The driver.sh script is extracted into /tmp/update.
	//  - /tmp/update/driver.sh /path/to/update.tar stage2 PRODUCT
	//  - The update tar is removed.
	//  - The device is rebooted.
	PackageFormatGeneric
)

// Parse is shorthand for Probe followed by Parse.
func Parse(z fs.FS, readTemp func(r io.Reader) (io.ReaderAt, error)) (Package, error) {
	pf, err := Probe(z)
	if err != nil {
		return Package{}, fmt.Errorf("probe update package: %w", err)
	}
	return pf.Parse(z, readTemp)
}

// Probe identifies the update format from the provided fs, which is generally a
// zip archive.
func Probe(z fs.FS) (PackageFormat, error) {
	if _, err := fs.Stat(z, "Kobo.tgz"); err == nil {
		return PackageFormatKobo, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return PackageFormatUnknown, err
	}
	if _, err := fs.Stat(z, "KoboRoot.tgz"); err == nil {
		return PackageFormatKoboRoot, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return PackageFormatUnknown, err
	}
	if _, err := fs.Stat(z, "update.tar"); err == nil {
		return PackageFormatGeneric, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return PackageFormatUnknown, err
	}
	return PackageFormatUnknown, nil
}

// Parse parses an update package from the provided fs, which is generally a zip
// archive.
//
// The provided readTemp function should read the contents of r into a temporary
// buffer, then return an [io.ReaderAt] reading from it. If the returned
// [io.ReaderAt] implements [io.Closer], it will be called automatically. If
// nil, [ReadTempDir] is used. Note that for generic update packages, multiple
// gigabytes of space may be used.
func (pf PackageFormat) Parse(z fs.FS, readTemp func(r io.Reader) (io.ReaderAt, error)) (Package, error) {
	if readTemp == nil {
		readTemp = ReadTempDir("")
	}
	p := Package{Format: pf}
	var err error
	switch p.Format {
	case PackageFormatKobo:
		err = p.parseKobo(z, readTemp)
	case PackageFormatKoboRoot:
		err = p.parseKoboRoot(z, readTemp)
	case PackageFormatGeneric:
		err = p.parseGeneric(z, readTemp)
	}
	return p, err
}

// String returns a human-readable string describing the package format.
func (pf PackageFormat) String() string {
	switch pf {
	case PackageFormatUnknown:
		return "unknown package"
	case PackageFormatKobo:
		return "Kobo.tgz package"
	case PackageFormatKoboRoot:
		return "KoboRoot.tgz package"
	case PackageFormatGeneric:
		return "generic update.tar package"
	default:
		return strconv.Itoa(int(pf))
	}
}

// String returns a human-readable string describing the package.
func (p Package) String() string {
	var b strings.Builder
	b.WriteString(p.Format.String())
	if p.Format != PackageFormatUnknown {
		b.WriteString(" {")
		if !p.Version.IsZero() {
			b.WriteString(" version=")
			b.WriteString(p.Version.String())
		}
		if !p.Date.IsZero() {
			b.WriteString(" date=")
			b.WriteString(p.Date.Format("2006-01-02"))
		}
		if p.Branch != "" {
			b.WriteString(" branch=")
			b.WriteString(p.Branch)
		}
		if p.Revision != "" {
			b.WriteString(" revision=")
			b.WriteString(p.Revision)
		}
		b.WriteString(" }")
	}
	return b.String()
}

func (p *Package) parseKobo(z fs.FS, _ func(r io.Reader) (io.ReaderAt, error)) error {
	r, err := z.Open("Kobo.tgz")
	if err != nil {
		return err
	}
	defer r.Close()

	rz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer rz.Close()

	if err := p.parse(func(fsDate func(t time.Time), push func(filename string, r io.Reader) error) error {
		fsDate(rz.ModTime)

		rzt := tar.NewReader(rz)
		for {
			h, err := rzt.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			name := "/usr/local/Kobo/" + strings.TrimPrefix(strings.TrimPrefix(h.Name, "./"), "/")
			switch name {
			case "/usr/local/Kobo/libnickel.so.1.0.0":
			case "/usr/local/Kobo/softwareversion":
			case "/usr/local/Kobo/branch":
			case "/usr/local/Kobo/revinfo":
			default:
				continue
			}
			if err := push(name, rzt); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if err := rz.Close(); err != nil {
		return err
	}
	return nil
}

func (p *Package) parseKoboRoot(z fs.FS, _ func(r io.Reader) (io.ReaderAt, error)) error {
	r, err := z.Open("KoboRoot.tgz")
	if err != nil {
		return err
	}
	defer r.Close()

	rz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer rz.Close()

	if err := p.parse(func(fsDate func(t time.Time), push func(filename string, r io.Reader) error) error {
		fsDate(rz.ModTime)

		rzt := tar.NewReader(rz)
		for {
			h, err := rzt.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			name := "/" + strings.TrimPrefix(strings.TrimPrefix(h.Name, "./"), "/")
			switch name {
			case "/usr/local/Kobo/libnickel.so.1.0.0":
			case "/usr/local/Kobo/softwareversion":
			case "/usr/local/Kobo/branch":
			case "/usr/local/Kobo/revinfo":
			default:
				continue
			}
			if err := push(name, rzt); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if err := rz.Close(); err != nil {
		return err
	}
	return nil
}

func (p *Package) parseGeneric(z fs.FS, readTemp func(r io.Reader) (io.ReaderAt, error)) error {
	r, err := z.Open("update.tar")
	if err != nil {
		return err
	}
	defer r.Close()

	rz, err := tarDecompressor(r)
	if err != nil {
		return err
	}

	rzt := tar.NewReader(rz)
	for {
		h, err := rzt.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		name := strings.TrimPrefix(strings.TrimPrefix(h.Name, "./"), "/")
		switch name {
		case "recoveryfs.img": // is probably a full recovery package
			if err := p.parse(func(fsDate func(t time.Time), push func(filename string, r io.Reader) error) error {
				recoveryfs, err := tarDecompressor(rzt)
				if err != nil {
					return fmt.Errorf("decompress recoveryfs: %w", err)
				}
				defer recoveryfs.Close()

				recoveryfsR, err := readTemp(recoveryfs)
				if err != nil {
					return fmt.Errorf("decompress recoveryfs: %w", err)
				}

				if err := recoveryfs.Close(); err != nil {
					return fmt.Errorf("decompress recoveryfs: %w", err)
				}

				if _, err := readExtSuperblock(recoveryfsR); err != nil {
					return fmt.Errorf("read recoveryfs superblock: %w", err)
				}

				recoveryfsT, ok := recoveryfsR.(*os.File)
				if !ok {
					return fmt.Errorf("read recoveryfs: readTemp must be backed by an *os.File")
				}

				if err := readExtFile(recoveryfsT, "/recovery/rootfs.ext4.zst", func(r io.Reader) error {
					rootfs, err := tarDecompressor(r)
					if err != nil {
						return fmt.Errorf("decompress rootfs: %w", err)
					}

					rootfsR, err := readTemp(rootfs)
					if err != nil {
						return fmt.Errorf("decompress rootfs: %w", err)
					}

					if err := rootfs.Close(); err != nil {
						return fmt.Errorf("decompress rootfs: %w", err)
					}

					if wtime, err := readExtSuperblock(rootfsR); err != nil {
						return fmt.Errorf("read rootfs superblock: %w", err)
					} else {
						fsDate(wtime)
					}

					rootfsT, ok := rootfsR.(*os.File)
					if !ok {
						return fmt.Errorf("read rootfs: readTemp must be backed by an *os.File")
					}

					for _, name := range []string{
						"/usr/local/Kobo/libnickel.so.1.0.0",
						"/usr/local/Kobo/softwareversion",
						"/usr/local/Kobo/branch",
						"/usr/local/Kobo/revinfo",
					} {
						if err := readExtFile(rootfsT, name, func(r io.Reader) error {
							return push(name, r)
						}); err != nil {
							return err
						}
					}

					return nil
				}); err != nil {
					return fmt.Errorf("read recoveryfs: %w", err)
				}
				return nil
			}); err != nil {
				return fmt.Errorf("process generic recovery package: %w", err)
			}
		}
	}

	if err := rz.Close(); err != nil {
		return err
	}
	return nil
}

func (p *Package) parse(handler func(fsDate func(t time.Time), push func(filename string, r io.Reader) error) error) error {
	var (
		verSW      Version
		verNickel  Version
		dateFS     time.Time
		dateNickel time.Time
	)
	if err := handler(
		func(t time.Time) {
			dateFS = t
		},
		func(filename string, r io.Reader) error {
			if err := func() error {
				buf, err := io.ReadAll(r)
				if err != nil {
					return err
				}
				switch filename {
				case "/usr/local/Kobo/libnickel.so.1.0.0":
					if i := bytes.Index(buf, []byte("Kobo Touch %2/%3")); i != -1 {
						if m := regexp.MustCompile(`[1234].[0-9]+\.[0-9]+`).FindAll(buf[i:i+200], -1); len(m) == 1 {
							if v, err := ParseVersion(string(m[0])); err == nil {
								verNickel = v
							}
						}
					}
					if i := bytes.Index(buf, []byte("%1/revinfo\x00")); i != -1 {
						if j := bytes.Index(buf[i:i+200], []byte("MMM d yyyy")); j != -1 {
							if m := regexp.MustCompile(`[A-z][a-z][a-z] [1-3]?[0-9] 2[0-9]{3}`).FindAll(buf[i:i+j], -1); len(m) == 1 {
								if v, err := time.ParseInLocation("Jan 2 2006", string(m[0]), time.UTC); err == nil {
									dateNickel = v
								}
							}
						}
					}
				case "/usr/local/Kobo/softwareversion": // provided since v5
					v, err := ParseVersion(strings.TrimSpace(string(buf)))
					if err != nil {
						return err
					}
					p.Version = v
					verSW = v
				case "/usr/local/Kobo/branch": // provided since v5
					p.Branch = strings.TrimSpace(string(buf))
				case "/usr/local/Kobo/revinfo":
					if f := strings.Fields(string(buf)); len(f) >= 1 {
						p.Revision = f[0]
					}
				default:
					panic(fmt.Errorf("wtf: unhandled filename %q", filename))
				}
				return nil
			}(); err != nil {
				return fmt.Errorf("process %s: %w", filename, err)
			}
			return nil
		},
	); err != nil {
		return err
	}
	if !verSW.IsZero() && !verNickel.IsZero() && verSW.Compare(verNickel) != 0 {
		return fmt.Errorf("mismatched softwareversion %s and libnickel version %s", verSW, verNickel)
	}
	if !verSW.IsZero() {
		p.Version = verSW
	} else if !verNickel.IsZero() {
		p.Version = verNickel
	}
	if dateNickel.Compare(dateFS) > 0 {
		p.Date = dateNickel
	} else {
		p.Date = dateFS
	}
	return nil
}
