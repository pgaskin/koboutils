package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
)

var ReadTempDirAuto = ReadTempDir

// ReadTempMem reads into an in-memory byte slice.
func ReadTempMem(r io.Reader) (io.ReaderAt, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf), nil
}

// ReadTempDir returns a function which reads into a temp file in d. If d is
// empty, the default temp folder is used. The file is closed and deleted when
// the [io.ReaderAt] is closed.
func ReadTempDir(d string) func(r io.Reader) (io.ReaderAt, error) {
	return func(r io.Reader) (io.ReaderAt, error) {
		if d == "" {
			d = os.TempDir()
		}
		if _, err := os.Stat(d); err != nil {
			return nil, fmt.Errorf("failed to access temp dir %q: %w", d, err)
		}
		f, err := os.CreateTemp(d, ".fwinfo-")
		if err != nil {
			return nil, fmt.Errorf("create temp file: %w", err)
		}
		if _, err := io.Copy(f, r); err != nil {
			return nil, fmt.Errorf("write temp file: %w", err)
		}
		return f, nil
	}
}

// tarDecompressor returns a reader decompressing r if known based on the magic
// used by [GNU tar].
//
// [GNU tar]:
// https://git.savannah.gnu.org/cgit/tar.git/tree/src/buffer.c?id=883f2e6dcaf87b4b449e55ed4f08dda1e701dae7#n295
func tarDecompressor(r io.Reader) (io.ReadCloser, error) {
	var magic string
	{
		tmp := make([]byte, 8)
		if n, err := r.Read(tmp); err != nil {
			return nil, err
		} else {
			magic = string(tmp[:n])
			r = io.MultiReader(bytes.NewReader(tmp), r)
		}
	}

	var (
		dec    string
		decRC  io.ReadCloser
		decErr error
	)
	switch {
	case strings.HasPrefix(magic, "\037\235"):
		dec, decRC = "compress", nil
	case strings.HasPrefix(magic, "\037\213"):
		var x *gzip.Reader
		x, decErr = gzip.NewReader(r)
		dec, decRC = "gzip", x
	case strings.HasPrefix(magic, "BZh"):
		dec, decRC = "bzip2", nil
	case strings.HasPrefix(magic, "LZIP"):
		dec, decRC = "lzip", nil
	case strings.HasPrefix(magic, "\xFFLZMA"), strings.HasPrefix(magic, "\x5d\x00\x00"):
		dec, decRC = "lzma", nil
	case strings.HasPrefix(magic, "\211LZO"):
		dec, decRC = "lzop", nil
	case strings.HasPrefix(magic, "\xFD7zXZ"):
		dec, decRC = "xz", nil
	case strings.HasPrefix(magic, "\x28\xB5\x2F\xFD"):
		var x *zstd.Decoder
		x, decErr = zstd.NewReader(r, zstd.WithDecoderLowmem(true))
		dec, decRC = "zstd", funcCloser{x, func() error {
			x.Close()
			return nil
		}}
	default:
		decRC = funcCloser{r, nil}
	}
	if decRC == nil {
		return nil, fmt.Errorf("unsupported decompressor %q (magic: %X)", dec, magic)
	}
	return decRC, decErr
}

// funcCloser wraps an io.Reader with a custom close function.
type funcCloser struct {
	r io.Reader
	c func() error
}

func (f funcCloser) Read(p []byte) (n int, err error) {
	return f.r.Read(p)
}

func (f funcCloser) Close() error {
	if f.c == nil {
		return nil
	}
	return f.c()
}

// readExtSuperblock reads the write time from the superblock of the provided
// ext4 filesystem.
func readExtSuperblock(r io.ReaderAt) (time.Time, error) {
	magic := make([]byte, 2)
	n, err := r.ReadAt(magic, 1024+0x38)
	if err == nil && n != len(magic) {
		err = io.ErrUnexpectedEOF
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("read ext magic: %w", err)
	}
	if !bytes.Equal(magic, []byte{0x53, 0xEF}) {
		return time.Time{}, fmt.Errorf("invalid ext magic %X", magic)
	}

	var wtime time.Time
	tmp := make([]byte, 4)
	n, err = r.ReadAt(tmp, 1024+0x30)
	if err == nil && n != len(tmp) {
		err = io.ErrUnexpectedEOF
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("read ext write time: %w", err)
	}
	if x := binary.LittleEndian.Uint32(tmp); x != 0 {
		wtime = time.Unix(int64(x), 0)
	}

	return wtime, nil
}

// readExtFile reads a file from the provided ext4 filesystem using 7-zip.
// Non-existent files will be empty.
func readExtFile(f *os.File, name string, fn func(r io.Reader) error) error {
	name = strings.TrimPrefix(name, "/")

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("7z.exe")
		cmd.Args = append(cmd.Args, "e", "-so", f.Name(), name)
	} else {
		cmd = exec.Command("7z")
		cmd.ExtraFiles = append(cmd.ExtraFiles, f)
		cmd.Args = append(cmd.Args, "e", "-so", "/proc/self/fd/3", name)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("extract %s with 7-zip: %w (stderr: %q)", name, err, strings.TrimSpace(stderr.String()))
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("extract %s with 7-zip: %w (stderr: %q)", name, err, strings.TrimSpace(stderr.String()))
	}

	if err := fn(stdout); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("extract %s with 7-zip: %w (stderr: %q)", name, err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

/* github.com/dsoprea/go-ext4 doesn't support hashed directories
// walkExtFS walks over the contents of an ext4 filesystem.
func walkExtFS(r io.ReaderAt, sbFn func(writeTime time.Time) error, walkFn func(name string, open func() (io.Reader, error)) error) error {
	return func() (perr error) {
		// that ext4 library panics for bad data structure
		defer func() {
			if err := recover(); err != nil {
				perr = fmt.Errorf("read fs: %v", err)
			}
		}()

		f := io.NewSectionReader(r, 0, 1<<63-1)

		if _, err := f.Seek(ext4.Superblock0Offset, io.SeekStart); err != nil {
			return fmt.Errorf("read superblock: %w", err)
		}

		sb, err := ext4.NewSuperblockWithReader(f)
		if err != nil {
			return fmt.Errorf("read superblock: %w", err)
		}

		if sbFn != nil {
			if err := sbFn(sb.WriteTime()); err != nil {
				return fmt.Errorf("process superblock: %w", err)
			}
		}

		if walkFn != nil {
			bgdl, err := ext4.NewBlockGroupDescriptorListWithReadSeeker(f, sb)
			if err != nil {
				return fmt.Errorf("read block group descriptor list: %w", err)
			}

			bgd, err := bgdl.GetWithAbsoluteInode(ext4.InodeRootDirectory)
			if err != nil {
				return fmt.Errorf("read root block group descriptor: %w", err)
			}

			dw, err := ext4.NewDirectoryWalk(f, bgd, ext4.InodeRootDirectory)
			if err != nil {
				return fmt.Errorf("create directory walker: %w", err)
			}

			for {
				fullPath, de, err := dw.Next()
				if err != nil {
					if err == io.EOF {
						break
					}
					return fmt.Errorf("walk fs: %w", err)
				}
				if err := walkFn(fullPath, func() (io.Reader, error) {
					r, err := func() (r io.Reader, perr error) {
						defer func() {
							if err := recover(); err != nil {
								perr = fmt.Errorf("%v", err)
							}
						}()

						abs := de.Data().Inode

						bgd, err := bgdl.GetWithAbsoluteInode(int(abs))
						if err != nil {
							return nil, fmt.Errorf("read block group descriptor for inode %d: %w", abs, err)
						}

						inode, err := ext4.NewInodeWithReadSeeker(bgd, f, int(abs))
						if err != nil {
							return nil, fmt.Errorf("read inode %d: %w", abs, err)
						}

						en := ext4.NewExtentNavigatorWithReadSeeker(f, inode)
						return ext4.NewInodeReader(en), nil
					}()
					if err != nil {
						err = fmt.Errorf("open %s: %w", fullPath, err)
					}
					return r, err
				}); err != nil {
					return fmt.Errorf("process %s: %w", fullPath, err)
				}
			}
		}
		return nil
	}()
}
*/
