package uu

import (
	"bufio"
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
	"testing/iotest"
)

type ByteSliceReader struct {
	remaining []byte
}

func (r *ByteSliceReader) Read(p []byte) (int, error) {
	if len(r.remaining) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.remaining)
	r.remaining = r.remaining[n:]
	return n, nil
}

func (r *ByteSliceReader) ReadByte() (byte, error) {
	if len(r.remaining) == 0 {
		return 0, io.EOF
	}

	var b byte
	b, r.remaining = r.remaining[0], r.remaining[1:]

	return b, nil
}

type newLineReader func([]byte) LineReader

func assertEOF(t *testing.T, reader LineReader) {
	l, e := reader.ReadLine()
	assert.Equal(t, io.EOF, e)
	assert.Nil(t, l)
}

func assertReads(t *testing.T, expected []byte, reader LineReader) {
	l, e := reader.ReadLine()
	assert.Nil(t, e)
	assert.Equal(t, expected, l)
}

func TestSliceLineReaderReadLine(t *testing.T) {
	testLineReaderReadLine(t, NewSliceLineReader)
}

func TestByteReaderLineReaderReadLine(t *testing.T) {
	f := func(in []byte) LineReader {
		return NewByteReaderLineReader(bytes.NewBuffer(in))
	}
	testLineReaderReadLine(t, f)
}

func TestByteReaderLineReaderReadLineError(t *testing.T) {
	expected := newError("Error test")
	lineReader := &byteReaderLineReader{reader: nil, err: expected}

	line, err := lineReader.ReadLine()
	assert.Nil(t, line)
	assert.Equal(t, expected, err)
}

func TestReaderLineReaderReadLine(t *testing.T) {
	f := func(in []byte) LineReader {
		return NewReaderLineReader(bytes.NewBuffer(in))
	}
	testLineReaderReadLine(t, f)
}

func TestReaderLineReaderReadLineError(t *testing.T) {
	expected := newError("Error test")
	lineReader := &readerLineReader{reader: nil, err: expected}

	line, err := lineReader.ReadLine()
	assert.Nil(t, line)
	assert.Equal(t, expected, err)
}

func TestBufioLineReaderReadLine(t *testing.T) {
	f := func(in []byte) LineReader {
		return NewBufioLineReader(bufio.NewReader(bytes.NewBuffer(in)))
	}
	testLineReaderReadLine(t, f)
}

func TestBufioLineReaderReadLine_error(t *testing.T) {
	expected := newError("Error test")
	lineReader := &bufioLineReader{reader: nil, err: expected}

	line, err := lineReader.ReadLine()
	assert.Nil(t, line)
	assert.Equal(t, expected, err)
}

func TestBufioLineReaderReadLine_partialLine(t *testing.T) {
	expected := []byte("a reasonably long line that could concievably come to be split if the buffer size is small enough")
	r := bufio.NewReaderSize(bytes.NewBuffer(expected), 20)
	reader := NewBufioLineReader(r)

	line, err := reader.ReadLine()
	assert.Nil(t, err)
	assert.Equal(t, expected, line)

	line, err = reader.ReadLine()
	assert.Nil(t, line)
	assert.Equal(t, io.EOF, err)
}

func TestBufioLineReaderReadLine_eofLine(t *testing.T) {
	expected := []byte("input")
	r := bufio.NewReader(iotest.DataErrReader(bytes.NewBuffer(expected)))
	reader := NewBufioLineReader(r)

	line, err := reader.ReadLine()
	assert.Nil(t, err)
	assert.Equal(t, expected, line)

	line, err = reader.ReadLine()
	assert.Nil(t, line)
	assert.Equal(t, io.EOF, err)
}

func testLineReaderReadLine(t *testing.T, newLineReader newLineReader) {
	reader := newLineReader([]byte("terminated\nunterminated"))
	assertReads(t, []byte("terminated"), reader)
	assertReads(t, []byte("unterminated"), reader)
	assertEOF(t, reader)

	reader = newLineReader([]byte("terminated\n"))
	assertReads(t, []byte("terminated"), reader)
	assertEOF(t, reader)

	reader = newLineReader([]byte("unterminated"))
	assertReads(t, []byte("unterminated"), reader)
	assertEOF(t, reader)

	reader = newLineReader([]byte(""))
	assertEOF(t, reader)

	reader = newLineReader([]byte("\n"))
	assertReads(t, []byte(""), reader)
	assertEOF(t, reader)
}
