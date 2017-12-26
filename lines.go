package uu

import (
	"bufio"
	"bytes"
	"io"
)

const lineLength = 64

// The LineReader is a common interface for reading newline separated lines from various sources
type LineReader interface {
	ReadLine() ([]byte, error)
}

type sliceLineReader struct {
	remaining []byte
}

type byteReaderLineReader struct {
	reader io.ByteReader
	err    error
}

type readerLineReader struct {
	reader  io.Reader
	err     error
	scratch []byte
}

type bufioLineReader struct {
	reader *bufio.Reader
	err    error
}

// NewSliceLineReader creates a LineReader for reading lines from a []byte slice
func NewSliceLineReader(bytes []byte) LineReader {
	if len(bytes) == 0 {
		return &sliceLineReader{remaining: nil}
	}
	return &sliceLineReader{remaining: bytes}
}

func (r *sliceLineReader) ReadLine() ([]byte, error) {
	i := bytes.IndexByte(r.remaining, '\n')
	if r.remaining == nil {
		return nil, io.EOF
	}
	var res []byte
	if i == -1 {
		res, r.remaining = r.remaining, nil
	} else if i+1 == len(r.remaining) {
		res, r.remaining = r.remaining[:i], nil
	} else {
		res, r.remaining = r.remaining[:i], r.remaining[i+1:]
	}
	return res, nil
}

// NewByteReaderLineReader creates a LineReader for reading lines from a io.ByteReader
func NewByteReaderLineReader(reader io.ByteReader) LineReader {
	return &byteReaderLineReader{reader: reader, err: nil}
}

func (r *byteReaderLineReader) ReadLine() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}

	var l []byte
	var b, err = r.reader.ReadByte()

	if err == nil {
		l = make([]byte, 0, lineLength)
		for err == nil && b != '\n' {
			l = append(l, b)
			b, err = r.reader.ReadByte()
		}
	}

	if err != nil {
		r.err = err
		r.reader = nil

		if l != nil && err == io.EOF {
			return l, nil
		}
		return nil, err
	}

	return l, nil
}

// NewReaderLineReader creates a LineReader for reading lines from a io.Reader
func NewReaderLineReader(reader io.Reader) LineReader {
	return &readerLineReader{reader: reader, err: nil, scratch: make([]byte, 1)}
}

func (r *readerLineReader) ReadLine() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}

	var l []byte
	var n, err = r.reader.Read(r.scratch)

	if n > 0 {
		l = make([]byte, 0, lineLength)
		for n > 0 && r.scratch[0] != '\n' {
			l = append(l, r.scratch[0])
			n, err = r.reader.Read(r.scratch)
		}
	}

	if err != nil {
		r.err = err
		r.reader = nil

		if l != nil && err == io.EOF {
			return l, nil
		}
		return nil, err
	}

	return l, nil
}

// NewBufioLineReader creates a LineReader for reading lines from a bufio.Reader
func NewBufioLineReader(reader *bufio.Reader) LineReader {
	return &bufioLineReader{reader: reader, err: nil}
}

func (r *bufioLineReader) ReadLine() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}

	var linePart, isPrefix, err = r.reader.ReadLine()
	var line []byte

	if err == nil {
		line = append(make([]byte, 0, lineLength), linePart...)
		for err == nil && isPrefix {
			linePart, isPrefix, err = r.reader.ReadLine()
			line = append(line, linePart...)
		}
	}

	if err != nil {
		r.err = err
		r.reader = nil

		if line != nil && err == io.EOF {
			return line, nil
		}
		return nil, err
	}

	return line, nil
}
