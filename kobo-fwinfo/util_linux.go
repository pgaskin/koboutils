package main

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/sys/unix"
)

func init() {
	ReadTempDirAuto = ReadTempDirAnon
}

// ReadTempDirAnon returns a function which reads into the provided directory
// with a file opened as O_TMPFILE. If d is empty, the default temp folder is
// used. The file is closed when the [io.ReaderAt] is closed.
func ReadTempDirAnon(d string) func(r io.Reader) (io.ReaderAt, error) {
	return func(r io.Reader) (io.ReaderAt, error) {
		if d == "" {
			d = os.TempDir()
		}
		if _, err := os.Stat(d); err != nil {
			return nil, fmt.Errorf("failed to access temp dir %q: %w", d, err)
		}
		f, err := os.OpenFile(d, unix.O_TMPFILE|unix.O_RDWR, 0666)
		if err != nil {
			return nil, fmt.Errorf("create temp file with O_TMPFILE: %w", err)
		}
		if _, err := io.Copy(f, r); err != nil {
			return nil, fmt.Errorf("write temp file: %w", err)
		}
		return f, nil
	}
}
