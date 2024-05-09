package internal

import (
	"bufio"
	"bytes"
	"debug/elf"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
)

// TODO: refactor this into a separate package, maybe even use for the next version of kobopatch

var ErrSegViolation = errors.New("segmentation violation")

type ELFMem struct {
	Elf *elf.File
}

// Index searches the mapped readable memory for the specified bytes starting at
// the specified virtual address, returning [math.MaxUint64] if no match is
// found.
func (m ELFMem) Index(b []byte, vaddr uint64) (i uint64, err error) {
	i, err = math.MaxUint64, nil
	m.Chunks(vaddr, func(vaddr, size uint64) bool {
		// read that chunk of memory
		buf := make([]byte, size)
		if _, err1 := io.ReadFull(m.ReaderAt(vaddr), buf); err1 != nil {
			i, err = math.MaxUint64, fmt.Errorf("read vaddr 0x%X - 0x%X (%d): %w", vaddr, vaddr+size, size, err1)
			return false
		}

		// search it
		if i1 := bytes.Index(buf, b); i1 != -1 {
			i, err = vaddr+uint64(i1), nil
			return false
		}

		// continue
		return true
	})
	return i, err
}

// Index is like Index, but searches for a struct using [binary.Write].
func (m ELFMem) IndexStruct(order binary.ByteOrder, data any, vaddr uint64) (i uint64, err error) {
	buf := bytes.NewBuffer(make([]byte, binary.Size(data))[:0])
	if err := binary.Write(buf, order, data); err != nil {
		return math.MaxUint64, err
	}
	return m.Index(buf.Bytes(), vaddr)
}

// EachIndex is like Index, but iterates over all occurrences (including
// overlaps) while fn returns true.
func (m ELFMem) EachIndex(b []byte, vaddr uint64, fn func(vaddr uint64) bool) error {
	var err error
	m.Chunks(vaddr, func(vaddr, size uint64) bool {
		// read that chunk of memory
		buf := make([]byte, size)
		if _, err1 := io.ReadFull(m.ReaderAt(vaddr), buf); err1 != nil {
			err = fmt.Errorf("read vaddr 0x%X - 0x%X (%d): %w", vaddr, vaddr+size, size, err1)
			return false
		}

		// search it
		for i1 := 0; i1 < len(buf); {
			i2 := bytes.Index(buf[i1:], b)
			if i2 == -1 {
				break
			}
			if !fn(vaddr + uint64(i1) + uint64(i2)) {
				return false
			}
			i1 += i2 + 1
		}

		// continue
		return true
	})
	return err
}

// Chunks iterates over non-overlapping chunks of mapped memory starting at
// vaddr while fn returns true.
func (m ELFMem) Chunks(vaddr uint64, fn func(vaddr, size uint64) bool) {
	for end := vaddr; ; {
		// find the start of the next non-zero-length segment with address >= end
		var off uint64 = math.MaxUint64
		for _, x := range m.Elf.Progs {
			if x.Vaddr >= end && x.Memsz > 0 {
				off = min(off, x.Vaddr)
			}
		}
		if off == math.MaxUint64 {
			break // no more segments to search
		}

		// find the next address not part of any segments
		end = off
		for _, x := range m.Elf.Progs {
			if x.Vaddr <= off {
				end = max(end, x.Vaddr+x.Memsz)
			}
		}
		if end == off {
			panic("wtf") // this should be impossible -- we already filter out zero-length segments
		}

		// handle it
		if !fn(off, end-off) {
			break
		}
	}
}

// ReaderAt returns a new reader starting at the specified virtual address.
func (m ELFMem) ReaderAt(vaddr uint64) io.Reader {
	return io.NewSectionReader(m, int64(vaddr), 1<<63-1)
}

// ReadStructAt is a wrapper for binary.Read.
func (m ELFMem) ReadStructAt(vaddr uint64, order binary.ByteOrder, data any) error {
	return binary.Read(m.ReaderAt(vaddr), binary.LittleEndian, data)
}

// ReadCString reads a null-terminated C string up to the specified length
// including the null terminator (0 for unlimited), returning io.EOF (and
// whatever was read so far) if no null terminator was found.
func (m ELFMem) ReadCString(vaddr uint64, max int) (string, error) {
	s, err := bufio.NewReader(io.NewSectionReader(m, int64(vaddr), int64(max))).ReadString(0)
	return strings.TrimSuffix(s, "\x00"), err
}

// ReadAt reads from the ELF file treating off as a virtual address.
func (m ELFMem) ReadAt(buf []byte, off int64) (n int, err error) {
	var (
		curOff = uint64(off)      // memory offset
		bufOff = uint64(0)        // read buffer offset
		bufLen = uint64(len(buf)) // read buffer length
	)
	for bufOff < bufLen {
		var (
			seg     *elf.Prog
			readOff = uint64(0)               // file offset from the start of seg
			readLen = uint64(bufLen - bufOff) // memory read length
			readNul = uint64(0)               // zero read length
		)

		// find the last segment which overlaps the current memory offset
		for _, x := range m.Elf.Progs {
			if x.Type != elf.PT_LOAD {
				continue
			}
			if x.Vaddr <= curOff && curOff < x.Vaddr+x.Memsz {
				readOff = curOff - x.Vaddr
				seg = x
				break
			}
		}

		// ensure we found a segment at the offset
		if seg == nil {
			return 0, fmt.Errorf("%w: no mapped segment at 0x%X + %d = 0x%X", ErrSegViolation, off, bufOff, curOff)
		}

		// limit the read length to the end of the segment
		if readOff+readLen > seg.Memsz {
			readLen = seg.Memsz - readOff
		}

		// limit the read length to the segment start offset of an overlapping segment, if any
		for _, x := range m.Elf.Progs {
			if x.Type != elf.PT_LOAD {
				continue
			}
			if curOff < x.Vaddr && curOff+readLen > x.Vaddr {
				readLen = x.Vaddr - curOff
				break
			}
		}

		// ensure we have permission to read the segment
		if seg.Flags&elf.PF_R == 0 {
			return 0, fmt.Errorf("%w: no permission to read mapped segment at 0x%X", ErrSegViolation, curOff)
		}

		// compute the amount of zeroes to read (i.e., when memory size is larger than file size)
		if readOff+readLen > seg.Filesz {
			readNul = readOff + readLen - seg.Filesz
		}

		// read the data
		n, err := seg.ReadAt(buf[bufOff:bufOff+readLen-readNul], int64(readOff))
		if n != int(readLen-readNul) {
			err = io.ErrUnexpectedEOF
		}
		if err != nil {
			return 0, err
		}
		for i := uint64(n); i < readLen; i++ {
			buf[i] = 0
		}

		// increment the current offset
		bufOff += readLen
		curOff += readLen
	}
	if bufOff > bufLen {
		panic("wtf") // should be impossible
	}
	if bufOff < bufLen {
		return 0, fmt.Errorf("%w: no mapped segment at 0x%X + %d = 0x%X", ErrSegViolation, off, bufOff, curOff)
	}
	return int(bufOff), nil
}
