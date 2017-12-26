package uu

import (
	"bytes"
	"io"
	"os"
	"strconv"
)

type encoding int

const (
	// uuEncoding File is UU encoded
	uuEncoding encoding = iota
	// base64Encoding File is Base64 encoded
	base64Encoding
)

type uuError struct {
	message string
}

func (e *uuError) Error() string {
	return e.message
}

// FileInfo is the exposes meta-data about the encoded data
type FileInfo struct {
	encoding encoding
	Name     string
	Mode     os.FileMode
}

// The Reader interface expose the UU functionality
type Reader interface {
	io.ByteReader
	io.Reader
	FileInfo() (*FileInfo, error)
}

type uuReader struct {
	reader  LineReader
	scratch []byte
	info    *FileInfo
	err     error
}

// NewReader creates a new Reader for decoding an UU encoded chunk from the provided LineReader
func NewReader(reader LineReader) Reader {
	return &uuReader{reader: reader, info: nil, err: nil, scratch: make([]byte, 0, 45)}
}

func (r *uuReader) nextOutByte() byte {
	b := r.scratch[0]
	r.scratch = r.scratch[1:]
	return b
}

func (r *uuReader) nextSlice(b []byte) int {
	n := copy(b, r.scratch)
	r.scratch = r.scratch[n:]
	return n
}

func (r *uuReader) ReadByte() (byte, error) {
	if len(r.scratch) > 0 {
		return r.nextOutByte(), nil
	}

	if r.info == nil {
		r.readInfo()
	}

	r.readLine()

	if r.err != nil {
		return 0, r.err
	}

	return r.nextOutByte(), nil
}

func (r *uuReader) Read(b []byte) (int, error) {
	if len(r.scratch) > 0 {
		return r.nextSlice(b), nil
	}

	if r.info == nil {
		r.readInfo()
	}

	r.readLine()

	if r.err != nil {
		return 0, r.err
	}

	return r.nextSlice(b), nil
}

func (r *uuReader) readLine() {
	if r.err != nil {
		return
	}
	line, err := r.reader.ReadLine()
	if err != nil {
		r.err = err
		return
	}

	r.scratch, err = parsePayloadLine(r.info, line, r.scratch)
	if err != nil {
		r.err = err
		if err == io.EOF {
			r.readEnd()
		}
	}
}

func (r *uuReader) readEnd() {
	line, err := r.reader.ReadLine()
	if err != nil {
		r.err = err
		return
	}

	err = parseEnd(r.info, line)
	if err != nil {
		r.err = err
		return
	}
}

func (r *uuReader) readInfo() {
	if r.err != nil {
		return
	}

	line, err := r.reader.ReadLine()
	if err != nil {
		r.err = err
		return
	}
	info, err := parseBegin(line)
	if err != nil {
		r.err = err
		return
	}

	r.info = info
}

func (r *uuReader) FileInfo() (*FileInfo, error) {
	if r.info == nil {
		r.readInfo()
	}

	return r.info, r.err
}

func newError(message string) *uuError {
	return &uuError{message: message}
}

func outLengthFromByte(b byte) (int, *uuError) {
	switch {
	case b == '`':
		return 0, nil
	case b >= ' ' && b <= 'M':
		return int(b) - ' ', nil
	default:
		return -1, newError("Invalid line length byte")
	}
}

func inLengthFromOutLength(outLength int) int {
	return ((outLength + 2) / 3) * 4
}

func fromEncoded(in byte) uint32 {
	return uint32(in-32) & 0x3f
}

func decode4to3(in []byte, out []byte) []byte {
	combined := fromEncoded(in[0])<<18 | fromEncoded(in[1])<<12 | fromEncoded(in[2])<<6 | fromEncoded(in[3])
	return append(out,
		byte(combined>>16),
		byte(combined>>8),
		byte(combined))
}

func decode4to2(in []byte, out []byte) []byte {
	combined := fromEncoded(in[0])<<18 | fromEncoded(in[1])<<12 | fromEncoded(in[2])<<6 | fromEncoded(in[3])
	return append(out,
		byte(combined>>16),
		byte(combined>>8))
}

func decode4to1(in []byte, out []byte) []byte {
	combined := fromEncoded(in[0])<<18 | fromEncoded(in[1])<<12 | fromEncoded(in[2])<<6 | fromEncoded(in[3])
	o2 := append(out, byte(combined>>16))
	return o2
}

func decodeBlocks(in []byte, out []byte, inLength int) ([]byte, int) {
	var i int
	for i = 0; i < inLength; i += 4 {
		out = decode4to3(in[i:i+4], out)
	}
	return out, i
}

func fileMode(mode []byte) (os.FileMode, error) {
	v, err := strconv.ParseUint(string(mode), 8, 32)
	if err != nil {
		return os.FileMode(0), newError("Failed to parse file mode: " + err.Error())
	}
	return os.FileMode(v), nil
}

func fileEncoding(begin []byte) (encoding, error) {
	switch string(begin) {
	case "begin":
		return uuEncoding, nil
	case "begin-base64":
		return base64Encoding, nil
	default:
		return 0, newError("Invalid header")
	}
}

func parseBegin(in []byte) (*FileInfo, error) {
	ibegin := bytes.IndexByte(in, ' ')
	if ibegin < 5 || len(in) < ibegin+1 {
		return nil, newError("Invalid header")
	}
	begin, tail := in[:ibegin], in[ibegin+1:]

	imode := bytes.IndexByte(tail, ' ')
	if imode < 1 || len(tail) < imode+1 {
		return nil, newError("Invalid header")
	}
	mode, tail := tail[:imode], tail[imode+1:]

	file := string(tail)

	encoding, err := fileEncoding(begin)
	if err != nil {
		return nil, err
	}
	fileMode, err := fileMode(mode)
	if err != nil {
		return nil, err
	}

	return &FileInfo{encoding: encoding, Mode: fileMode, Name: file}, nil
}

func parseEnd(fileInfo *FileInfo, in []byte) error {
	if !bytes.Equal(in, endMarker(fileInfo)) {
		return newError("Invalid trailer")
	}
	return nil
}

func endMarker(fileInfo *FileInfo) []byte {
	switch fileInfo.encoding {
	case uuEncoding:
		return []byte("end")
	case base64Encoding:
		return []byte("====")
	}
	panic("Invalid encoding")
}

func parsePayloadLine(fileInfo *FileInfo, in []byte, out []byte) ([]byte, error) {
	outLength, err := outLengthFromByte(in[0])
	if err != nil {
		return nil, err
	}
	if outLength == 0 {
		return nil, io.EOF
	}

	inLength := inLengthFromOutLength(outLength)
	if len(in) < inLength+1 { // + 1 for length byte
		return nil, newError("Input line too short")
	}

	payload := in[1:]
	var i int

	switch outLength % 3 {
	case 0:
		out, _ = decodeBlocks(payload, out, inLength)
	case 1:
		out, i = decodeBlocks(payload, out, inLength-4)
		out = decode4to1(payload[i:], out)
	case 2:
		out, i = decodeBlocks(payload, out, inLength-4)
		out = decode4to2(payload[i:], out)
	}

	return out, nil
}
