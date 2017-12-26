package uu_test

import (
	"bytes"
	"bufio"
	"io/ioutil"
	"uu"
)

const data = `begin 644 hello.txt
,2&5L;&\@5V]R;&0*
`+ "`" + `
end
`


// Hello
func ExampleReader_bufio() {
	reader := bufio.NewReader(bytes.NewBuffer([]byte(data)))

	uureader := uu.NewReader(uu.NewBufReaderLineReader(reader))
	fileInfo, err := uureader.FileInfo()
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

