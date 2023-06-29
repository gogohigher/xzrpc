package gob

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"github.com/gogohigher/xzrpc/pkg/protocol"
	"github.com/gogohigher/xzrpc/pkg/traffic"
	"io"
)

var _ protocol.Protocol = (*gobProtocol)(nil)

type gobProtocol struct {
	conn    io.ReadWriteCloser
	buf     *bufio.Writer
	encoder *gob.Encoder
	decoder *gob.Decoder
}

func NewGobProtocol(rw io.ReadWriteCloser) protocol.Protocol {
	buf := bufio.NewWriter(rw)
	return &gobProtocol{
		conn:    rw,
		buf:     buf,
		encoder: gob.NewEncoder(buf),
		decoder: gob.NewDecoder(rw),
	}
}

// 协议内容
func (gp *gobProtocol) GetContent() *protocol.Content {
	c := &protocol.Content{
		Id:   2,
		Name: "gob",
	}
	return c
}

// 封包，即将message写入connection
func (gp *gobProtocol) Pack(msg traffic.Message) error {
	var err error
	defer func() {
		_ = gp.buf.Flush() // buf在NewEncoder的时候，作为传输传入
		if err != nil {
			_ = gp.Close()
		}
	}()
	if err := gp.encoder.Encode(msg.Header()); err != nil {
		fmt.Println("failed to encode header: ", err)
		return err
	}
	if err := gp.encoder.Encode(msg.Body()); err != nil {
		fmt.Println("failed to encode body: ", err)
		return err
	}
	return nil
}

// 拆包，从connection中读取消息，读到message中
func (gp *gobProtocol) UnPack(msg traffic.Message, f func(header traffic.Header) error) error {
	if err := gp.decoder.Decode(msg.Header()); err != nil {
		return err
	}
	return gp.decoder.Decode(msg.Body())
}

func (gp *gobProtocol) Close() error {
	return gp.conn.Close()
}
