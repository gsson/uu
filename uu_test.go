package uu

import (
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestLineBytes(t *testing.T) {
	assertLineBytes(t, 'M', 45)
	assertLineBytes(t, '!', 1)
	assertLineBytes(t, '`', 0)
	assertLineBytes(t, '`', 0)
	assertLineBytesError(t, 'N')
	assertLineBytesError(t, 0)
	assertLineBytesError(t, 255)
}

func TestInputLength(t *testing.T) {
	assert.Equal(t, inLengthFromOutLength(0), 0)
	assert.Equal(t, inLengthFromOutLength(1), 4)
	assert.Equal(t, inLengthFromOutLength(2), 4)
	assert.Equal(t, inLengthFromOutLength(3), 4)
	assert.Equal(t, inLengthFromOutLength(4), 8)
}

func TestDecode4to3(t *testing.T) {
	out := make([]byte, 0, 3)
	out = decode4to3([]byte("0V%T"), out)
	assert.Equal(t, []byte("Cat"), out)
}

func TestDecode4to2(t *testing.T) {
	out := make([]byte, 0, 2)
	out = decode4to2([]byte("0V%T"), out)
	assert.Equal(t, []byte("Ca"), out)
}

func TestDecode4to1(t *testing.T) {
	out := make([]byte, 0, 1)
	out = decode4to1([]byte("0V%T"), out)
	assert.Equal(t, []byte("C"), out)
}

func TestParsePayloadLine(t *testing.T) {
	assertPayloadLineEOF(t, "`\n")
	assertPayloadLineParsed(t, "!80``\n", "a")
	assertPayloadLineParsed(t, "\"86(`\n", "ab")
	assertPayloadLineParsed(t, "#86)C\n", "abc")
	assertPayloadLineParsed(t, "$86)C9```\n", "abcd")
	assertPayloadLineParsed(t, "%86)C9&4`\n", "abcde")
	assertPayloadLineParsed(t, "&86)C9&5F\n", "abcdef")
	assertPayloadLineParsed(t, "::'1T<#HO+W=W=RYW:6MI<&5D:6$N;W)G#0H`\n", "http://www.wikipedia.org\r\n")
}

func TestParseBeginLine(t *testing.T) {
	assertBeginLineParsed(t, "begin 000 hello.txt", FileInfo{encoding: uuEncoding, Mode: os.FileMode(0), Name: "hello.txt"})
	assertBeginLineParsed(t, "begin-base64 000 hello.txt", FileInfo{encoding: base64Encoding, Mode: os.FileMode(0), Name: "hello.txt"})

	assertBeginLineFails(t, "begin-base63 000 hello.txt", "Invalid header")
	assertBeginLineFails(t, "", "Invalid header")
	assertBeginLineFails(t, "begi", "Invalid header")
	assertBeginLineFails(t, "begin aaa hello.txt", "Failed to parse file mode: strconv.ParseUint: parsing \"aaa\": invalid syntax")
	assertBeginLineFails(t, "begin 000", "Invalid header")
}

func TestParseEndLine(t *testing.T) {
	assertEndLineParsed(t, "end", FileInfo{encoding: uuEncoding, Mode: os.FileMode(0), Name: "hello.txt"})
	assertEndLineParsed(t, "====", FileInfo{encoding: base64Encoding, Mode: os.FileMode(0), Name: "hello.txt"})

	assertEndLineFails(t, "====", FileInfo{encoding: uuEncoding, Mode: os.FileMode(0), Name: "hello.txt"})
	assertEndLineFails(t, "end", FileInfo{encoding: base64Encoding, Mode: os.FileMode(0), Name: "hello.txt"})
	assertEndLineFails(t, "fnord", FileInfo{encoding: uuEncoding, Mode: os.FileMode(0), Name: "hello.txt"})
	assertEndLineFails(t, "fnord", FileInfo{encoding: base64Encoding, Mode: os.FileMode(0), Name: "hello.txt"})
}

func TestDecodeFile(t *testing.T) {
	assertDecodes(t, "test.uu", 0644, "test.bin")
}

func decodeWithReadByte(t *testing.T, reader io.ByteReader) []byte {
	contents := make([]byte, 0)
	b, err := reader.ReadByte()
	for err == nil {
		contents = append(contents, b)
		b, err = reader.ReadByte()
	}
	assert.EqualError(t, err, "EOF")

	return contents
}

func decodeWithReader(t *testing.T, reader io.Reader) []byte {
	contents, err := ioutil.ReadAll(reader)

	assert.Nil(t, err)

	return contents
}

func openTestData(uuFileName string) Reader {
	f, err := os.Open("testdata/" + uuFileName)
	if err != nil {
		panic(err)
	}

	return NewReader(NewReaderLineReader(f))
}

func assertDecodes(t *testing.T, uuFileName string, mode os.FileMode, fileName string) {
	expected, err := ioutil.ReadFile("testdata/" + fileName)
	if err != nil {
		panic(err)
	}

	r := openTestData(uuFileName)
	fileInfo, err := r.FileInfo()
	assert.Nil(t, err)
	assert.Equal(t, fileName, fileInfo.Name)
	assert.Equal(t, mode, fileInfo.Mode)
	assert.Equal(t, uuEncoding, fileInfo.encoding)

	contents := decodeWithReadByte(t, r)
	assert.Equal(t, expected, contents)

	r = openTestData(uuFileName)

	contents = decodeWithReader(t, r)
	assert.Equal(t, expected, contents)
}

func assertBeginLineParsed(t *testing.T, in string, expected FileInfo) {
	fileInfo, err := parseBegin([]byte(in))
	assert.Nil(t, err)
	assert.Equal(t, &expected, fileInfo)
}

func assertBeginLineFails(t *testing.T, in string, message string) {
	fileInfo, err := parseBegin([]byte(in))
	assert.Nil(t, fileInfo)
	assert.EqualError(t, err, message)
}

func assertEndLineParsed(t *testing.T, in string, fileInfo FileInfo) {
	err := parseEnd(&fileInfo, []byte(in))
	assert.Nil(t, err)
}

func assertEndLineFails(t *testing.T, in string, expected FileInfo) {
	err := parseEnd(&expected, []byte(in))
	assert.EqualError(t, err, "Invalid trailer")
}

func assertPayloadLineParsed(t *testing.T, in string, expected string) {
	fileInfo := FileInfo{encoding: uuEncoding, Mode: os.FileMode(0), Name: "hello.txt"}
	out := make([]byte, 0, len(expected))
	out, err := parsePayloadLine(&fileInfo, []byte(in), out)
	assert.Nil(t, err)
	assert.Equal(t, []byte(expected), out)
}

func assertPayloadLineEOF(t *testing.T, in string) {
	fileInfo := FileInfo{encoding: uuEncoding, Mode: os.FileMode(0), Name: "hello.txt"}
	out := make([]byte, 0)
	out, err := parsePayloadLine(&fileInfo, []byte(in), out)
	assert.Nil(t, out)
	assert.Equal(t, err, io.EOF)
}

func assertLineBytes(t *testing.T, in byte, expected int) {
	var v, err = outLengthFromByte(in)
	assert.Nil(t, err)
	assert.Equal(t, v, expected)
}

func assertLineBytesError(t *testing.T, in byte) {
	_, err := outLengthFromByte(in)
	assert.Error(t, err)
}
