package compressor

import (
	"fmt"
	"testing"
)

var zipBytes []byte

func TestZip(t *testing.T) {
	var data = []byte{'a', 'b', 'c', 'a', 'b', 'c', 'a', 'a', 'a'}
	fmt.Println("data length: ", len(data))

	var gzip = &GzipCompressor{}
	zipBytes, _ = gzip.Zip(data)

	fmt.Println("data: ", string(zipBytes))

	fmt.Println("zipBytes: ", zipBytes)

	UnZip()
}

func UnZip() {
	var gzip = &GzipCompressor{}
	unzipBytes, _ := gzip.UnZip(zipBytes)
	fmt.Println("data: ", string(unzipBytes))
}
