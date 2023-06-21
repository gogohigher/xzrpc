package codec

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"github.com/gogohigher/xzrpc/pkg/traffic"
	"io"
)

// 检测GobCodec是否实现了Codec接口
var _ Codec = (*GobCodec)(nil)

// GobCodec 默认自带的编解码器
type GobCodec struct {
	conn    io.ReadWriteCloser
	buf     *bufio.Writer
	encoder *gob.Encoder
	decoder *gob.Decoder
}

// NewGobCodec 创建GobCodec对象
func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn:    conn,
		buf:     buf,
		encoder: gob.NewEncoder(buf),
		decoder: gob.NewDecoder(conn),
	}
}

func (gc *GobCodec) ReadHeader(header *traffic.Header) error {

	return gc.decoder.Decode(header)
}

func (gc *GobCodec) ReadBody(body interface{}) error {
	return gc.decoder.Decode(body)
}

// 1. 先encode header；2. 再encode body
func (gc *GobCodec) Write(header *traffic.Header, body interface{}) (err error) {
	defer func() {
		_ = gc.buf.Flush() // buf在NewEncoder的时候，作为传输传入
		if err != nil {
			_ = gc.Close()
		}
	}()
	if err := gc.encoder.Encode(header); err != nil {
		fmt.Println("failed to encode header: ", err)
		return err
	}
	if err := gc.encoder.Encode(body); err != nil {
		fmt.Println("failed to encode header: ", err)
		return err
	}
	return nil
}

func (gc *GobCodec) Close() error {
	return gc.conn.Close()
}
