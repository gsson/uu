# UU

[![Build Status](https://travis-ci.org/gsson/uu.svg)](https://travis-ci.org/gsson/uu) [![Build Status](https://goreportcard.com/badge/github.com/gsson/uu)](https://goreportcard.com/report/github.com/gsson/uu)

UU decoder written in Go because why not.

## Usage

```
package main

import (
	"github.com/gsson/uu"
	"io/ioutil"
	"os"
)

func main() {
	f, err := os.Open("uuencoded.uu")
	if err != nil {
		panic(err)
	}

	reader := uu.NewReader(uu.NewReaderLineReader(f))
	fileInfo, err := reader.FileInfo()
	if err != nil {
		panic(err)
	}

	println(fileInfo.Name)

	uudecoded, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	println(string(uudecoded))
}
```

## Notes

Currently the `uu.Reader` only supports the classic UU format (no Base64)

`uu.Reader` implements the `io.ByteReader` and `io.Reader` interfaces.

There are `uu.LineReader` implementations for reading from `io.ByteReader` (`NewByteReaderLineReader`), `io.Reader` (`NewReaderLineReader`), `bufio.Reader` (`NewBufReaderLineReader`) and `[]byte` slices (`NewSliceLineReader`).

Note that the `io.ByteReader` and `io.Reader` `uu.LineReader` implementations might be slow as they read byte-by-byte to prevent over-reading at the end of the encoded file since the input could contain multiple entries.

If this is an issue, try using the `bufio.Reader` implementation.
