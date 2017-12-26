package uu_test

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/gsson/uu"
	"io/ioutil"
)

// Read single chunk from []byte slice
func Example() {
	input := "begin 644 hello.txt\n" +
		",2&5L;&\\@5V]R;&0*\n" +
		"`\n" +
		"end\n"

	uureader := uu.NewReader(uu.NewSliceLineReader([]byte(input)))
	fileInfo, err := uureader.FileInfo()
	if err != nil {
		panic(err)
	}

	contents, err := ioutil.ReadAll(uureader)
	if err != nil {
		panic(err)
	}

	fmt.Println(fileInfo.Name)
	fmt.Print(string(contents))

	// Output:
	// hello.txt
	// Hello World
}

// Read single chunk from bufio.Reader
func ExampleReader_singleBufio() {
	input := "begin 644 hello.txt\n" +
		",2&5L;&\\@5V]R;&0*\n" +
		"`\n" +
		"end\n"
	reader := bufio.NewReader(bytes.NewBufferString(input))

	uureader := uu.NewReader(uu.NewBufioLineReader(reader))
	fileInfo, err := uureader.FileInfo()
	if err != nil {
		panic(err)
	}

	contents, err := ioutil.ReadAll(uureader)
	if err != nil {
		panic(err)
	}

	fmt.Println(fileInfo.Name)
	fmt.Print(string(contents))

	// Output:
	// hello.txt
	// Hello World
}

// Read two consecutive files from the same bufio.Reader
func ExampleReader_multipleBufio() {
	input := "begin 644 hello.txt\n" +
		",2&5L;&\\@5V]R;&0*\n" +
		"`\n" +
		"end\n" +
		"begin 644 lorem.txt\n" +
		"M3&]R96T@:7!S=6T@9&]L;W(@<VET(&%M970L(&-O;G-E8W1E='5R(&%D:7!I\n" +
		"M<V-I;F<@96QI=\"P@<V5D(&1O(&5I=7-M;V0@=&5M<&]R(&EN8VED:61U;G0@\n" +
		"B=70@;&%B;W)E(&5T(&1O;&]R92!M86=N82!A;&EQ=6$N\"@``\n" +
		"`\n" +
		"end\n"
	reader := bufio.NewReader(bytes.NewBufferString(input))

	uureader := uu.NewReader(uu.NewBufioLineReader(reader))
	fileInfo, err := uureader.FileInfo()
	if err != nil {
		panic(err)
	}

	contents, err := ioutil.ReadAll(uureader)
	if err != nil {
		panic(err)
	}

	fmt.Println(fileInfo.Name)
	fmt.Print(string(contents))

	uureader = uu.NewReader(uu.NewBufioLineReader(reader))
	fileInfo, err = uureader.FileInfo()
	if err != nil {
		panic(err)
	}

	contents, err = ioutil.ReadAll(uureader)
	if err != nil {
		panic(err)
	}

	fmt.Println(fileInfo.Name)
	fmt.Print(string(contents))

	// Output:
	// hello.txt
	// Hello World
	// lorem.txt
	// Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
}
