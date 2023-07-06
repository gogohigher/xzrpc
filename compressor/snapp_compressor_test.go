package compressor

import (
	"fmt"
	"testing"
)

func TestSnappyZip(t *testing.T) {
	var data = []byte{'a', 'b', 'c', 'a', 'b', 'c', 'a', 'a', 'a'}
	fmt.Println("data: ", string(data), ", and length is ", len(data))

	var snappyC = &SnappyCompressor{}
	snappyZipBytes, _ := snappyC.Zip(data)

	fmt.Println("snappyBytes: ", string(snappyZipBytes), ", and length is ", len(snappyZipBytes))

	unzipBytes, _ := snappyC.UnZip(snappyZipBytes)
	fmt.Println("unzip snappyBytes: ", string(unzipBytes))
}
