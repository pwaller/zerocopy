// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

// zerocopy.Reader gives a reading interface for a byte slice or a file, which
// doesn't make a copy of the underlying byte slice.
// To achieve this, a new "Read" interface is required, where the data storage
// is not specified by the caller, but by the callee. This is so that the
// implementation is at liberty to return a byte slice
package zerocopy

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"unsafe"

	"github.com/edsrzf/mmap-go"
)

// Construct a new Zero-Copy reader on `r`.
// `r` must be an *os.File or a *bytes.Reader, otherwise an error is returned.
func NewReader(r io.Reader) (ReadCloser, error) {
	switch r := r.(type) {
	case *bytes.Reader:
		return newBytesReader(r)

	case *os.File:
		return newMmapReader(r)

	default:
		// Note: I prefer to return an error for the moment, rather than
		// return a "fake" zero-copy reader which actually does a copy.
		return nil, fmt.Errorf("No zero-copy reader implementation available for %T", r)
	}
}

// Helper to construct a zero-copy reader directly from a byte slice
func NewReaderFromBytes(b []byte) (ReadCloser, error) {
	return NewReader(bytes.NewReader(b))
}

// The zero-copy Reader interface.
// It has a different Read() method than the usual because the semantics
// are different.
type Reader interface {
	// Read `size` bytes from the underlying stream, and return a byte slice
	// to those bytes. The returned []byte is a slice which references the
	// underlying bytes, and must not be written to.
	Read(size uint64) ([]byte, error)
}

type ReadCloser interface {
	Reader
	Close() error
}

// bytesReader provides zero-copy Reads on an existing bytes.Reader by futzing
// around with its internal. Caveat Emptor.
type bytesReader struct {
	b *bytes.Reader
	s []byte
	i *int
}

// Construct a new zerocopy.Reader
func newBytesReader(r *bytes.Reader) (*bytesReader, error) {
	old := reflect.ValueOf(r).Elem()
	newSlice := copyPrivateByteSlice(old.FieldByName("s"))
	iptr := (*int)(unsafe.Pointer(old.FieldByName("i").UnsafeAddr()))
	return &bytesReader{r, newSlice, iptr}, nil

}

func (r *bytesReader) Read(size uint64) ([]byte, error) {
	if size == 0 {
		return nil, nil
	}
	if *r.i >= len(r.s) {
		return nil, io.EOF
	}
	// TODO(pwaller): Ideally, we'd set this on the underlying reader, but
	// 				  we assume that we own it for now.
	// r.prevRune = -1
	result := r.s[*r.i : *r.i+int(size)]
	*r.i += int(size)
	return result, nil
}

func (r *bytesReader) Close() error {
	return nil
}

// The mmapReader implementation works by using the bytesReader implemenation
// on an mmap'ed byte-array
type mmapReader struct {
	Reader
	mmap.MMap
}

func newMmapReader(r *os.File) (*mmapReader, error) {
	mapping, err := mmap.Map(r, mmap.RDONLY, 0)
	if err != nil {
		return nil, err
	}

	underlying, err := NewReaderFromBytes(mapping)
	if err != nil {
		return nil, err
	}

	return &mmapReader{
		underlying,
		mapping,
	}, nil
}

func (m *mmapReader) Close() error {
	return m.Unmap()
}

// Obtain a copy of a private byte slice
func copyPrivateByteSlice(value reflect.Value) []byte {
	newSlice := []byte{}
	_newSlice := (*reflect.SliceHeader)(unsafe.Pointer(&newSlice))
	_origSlice := (*reflect.SliceHeader)(unsafe.Pointer(value.UnsafeAddr()))
	_newSlice.Data = _origSlice.Data
	_newSlice.Len = _origSlice.Len
	_newSlice.Cap = _origSlice.Cap
	return newSlice
}
