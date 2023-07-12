package compressor

import (
	"bytes"
	"github.com/golang/snappy"
	"io"
)

var _ Compressor = (*SnappyCompressor)(nil)

type SnappyCompressor struct {
}

func (c *SnappyCompressor) Zip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := snappy.NewBufferedWriter(&buf)
	defer func() {
		_ = w.Close()
	}()
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *SnappyCompressor) UnZip(data []byte) ([]byte, error) {
	r := snappy.NewReader(bytes.NewReader(data))
	// 内部做了处理，err == EOF时，err = nil
	res, err := io.ReadAll(r)
	if err != nil {
		// 需要排除ErrUnexpectedEOF错误
		if err != io.ErrUnexpectedEOF {
			return nil, err
		}
	}
	return res, nil
}
