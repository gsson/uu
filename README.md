# UU

[![Build Status](https://travis-ci.org/gsson/uu.svg)](https://travis-ci.org/gsson/uu) [![Go Report Card](https://goreportcard.com/badge/github.com/gsson/uu)](https://goreportcard.com/report/github.com/gsson/uu) [![License](https://img.shields.io/github/license/gsson/uu.svg?maxAge=2592000)](https://github.com/gsson/uu/blob/master/LICENSE) [![Documentation](https://godoc.org/github.com/gsson/uu?status.svg)](http://godoc.org/github.com/gsson/uu)
UU decoder written in Go because why not.

## Usage

See the [Documentation](http://godoc.org/github.com/gsson/uu)

## Notes

Currently the `uu.Reader` only supports the classic UU format (no Base64)

`uu.Reader` implements the `io.ByteReader` and `io.Reader` interfaces.

There are `uu.LineReader` implementations for reading from `io.ByteReader` (`NewByteReaderLineReader`), `io.Reader` (`NewReaderLineReader`), `bufio.Reader` (`NewBufioLineReader`) and `[]byte` slices (`NewSliceLineReader`).

Note that the `io.ByteReader` and `io.Reader` `uu.LineReader` implementations might be slow as they read byte-by-byte to prevent over-reading at the end of the encoded file since the input could contain multiple entries.

If this is an issue, try using the `bufio.Reader` implementation.
