package raw

// raw_protocol对应的消息格式如下：
// message length		占4个字节
// 协议Id				占1个字节
// 消息序列号长度 			占1个字节
// 序列号内容
// 消息类型 				占1个字节
// serviceMethod 长度 	占1个字节
// serviceMethod 内容
// body codec id  		占1个字节
// body

/*
	raw_protocol对应的消息格式如下：
   +----------------+-----------+---------------+-----------+-----------+-------------------+-----------------------+---------------+-------+
   | message length | 协议Id 	| 消息序列号长度 	| 序列号内容 	| 消息类型 	| serviceMethod 长度 | serviceMethod 内容 	| body codec id | body  |
   +----------------------------------------------------------------------------------------------------------------------------------------+
   | 占4个字节 		| 占1个字节 	| 占1个字节 		| xxx 		| 占1个字节 	| 占1个字节 			| xxx 					| 占1个字节 		| xxx 	|
   +----------------+-----------+---------------+-----------+-----------+-------------------+-----------------------+---------------+-------+
*/

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gogohigher/xzrpc/internal/xzbufio"
	"github.com/gogohigher/xzrpc/protocol"
	"github.com/gogohigher/xzrpc/traffic"
	"io"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
)

var _ protocol.Protocol = (*rawProtocol)(nil)

type rawProtocol struct {
	r  io.Reader
	w  io.Writer
	mu sync.Mutex
}

func NewRawProtocol(rw io.ReadWriteCloser) protocol.Protocol {
	return &rawProtocol{
		r: rw,
		w: rw,
	}
}

// GetContent 协议内容
func (raw *rawProtocol) GetContent() *protocol.Content {
	c := &protocol.Content{
		Id:   1,
		Name: "raw",
	}
	return c
}

// 封包，即将message写入connection
func (raw *rawProtocol) Pack(msg traffic.Message) error {
	// TODO 使用sync.Pool进行优化；缓冲区大小取什么值比较好？
	//buf := make([]byte, 0, 4096)
	buf := xzbufio.NewWriter(raw.w) // TODO 内部默认是4096，也可以修改默认大小
	// 消息长度 | 占据4个字节
	err := binary.Write(buf, binary.BigEndian, uint32(0))
	if err != nil {
		return fmt.Errorf("write message err: %v", err.Error())
	}
	err = raw.WriteHeader(buf, msg)
	if err != nil {
		return err
	}
	err = raw.WriteBody(buf, msg)
	if err != nil {
		return err
	}
	//bufSize := buf.Size() - buf.Available()
	length := buf.Len()
	log.Printf("raw-protocol | Pack buf length: %d\n", length)

	// 下面的方法：以大端的方式，写入二进制。效果是追加到后面。这里应该需要将整个消息的大小写到前4位
	//if err := binary.Write(buf, binary.BigEndian, uint32(bufSize)); err != nil {
	//	return err
	//}
	binary.BigEndian.PutUint32(buf.B(), uint32(length))

	log.Println("Pack buf: ", buf.B()[:length])
	log.Println("Pack buf: ", string(buf.B()[:length]))

	_ = buf.Flush()
	return nil
}

// TODO 处理error
func (raw *rawProtocol) WriteHeader(buf *xzbufio.Writer, msg traffic.Message) error {
	// 协议Id | 占1个字节
	_ = buf.WriteByte(raw.GetContent().Id)
	// 消息序列号长度 | 占1个字节
	seqStr := strconv.FormatInt(int64(msg.GetSeq()), 36)
	_ = buf.WriteByte(byte(len(seqStr)))
	// 序列号内容
	_, _ = buf.Write([]byte(seqStr))
	// 消息类型 | 占1个字节
	_ = buf.WriteByte(msg.Action())
	// serviceMethod 长度 | 占1个字节
	sm := strings.TrimSpace(msg.ServiceMethod())
	if len(sm) > math.MaxUint8 {
		return errors.New("raw-protocol | length of service method must not be more than 255")
	}
	_ = buf.WriteByte(byte(len(sm)))
	// serviceMethod 内容
	_, _ = buf.Write([]byte(sm))
	// body codec id | 占1个字节
	_ = buf.WriteByte(msg.Codec())
	return nil
}

func (raw *rawProtocol) WriteBody(buf *xzbufio.Writer, msg traffic.Message) error {
	// body
	bytes, err := msg.MarshalBody()
	if err != nil {
		log.Println("raw-protocol | WriteBody | failed to MarshalBody: ", err)
		return err
	}
	_, err = buf.Write(bytes)
	if err != nil {
		log.Println("raw-protocol | WriteBody | failed to Write to buf: ", err)
		return err
	}
	return nil
}

// UnPack 拆包，从connection中读取消息，读到message中
// 1. 根据前4位，获取到整个byte数组的大小
// 2. 根据大小，创建byte数组，然后将所有数据读到这个byte数组中
func (raw *rawProtocol) UnPack(msg traffic.Message, f func(header traffic.Header) error) error {
	raw.mu.Lock()
	defer raw.mu.Unlock()

	buf := make([]byte, 4)
	n, err := io.ReadFull(raw.r, buf)
	if err != nil {
		return err
	}
	if n != 4 {
		return errors.New("buf length error")
	}
	wholeByteSize := binary.BigEndian.Uint32(buf)
	log.Println("raw-protocol | UnPack wholeByteSize: ", wholeByteSize)

	msgSize := wholeByteSize - 4
	buf = make([]byte, msgSize)
	if _, err := io.ReadFull(raw.r, buf); err != nil {
		return err
	}

	// parse buf
	log.Println("UnPack buf: ", buf)
	log.Println("UnPack buf: ", string(buf))
	data, err := raw.ReadHeader(buf, msg)
	if err != nil {
		log.Println("raw-protocol | UnPack | failed to ReadHeader: ", err)
		return err
	}
	if f != nil {
		if err := f(msg.Header()); err != nil {
			return err
		}

	}
	// TODO 这里需要处理body的类型
	// 外部解决
	if err := raw.ReadBody(data, msg); err != nil {
		log.Println("raw-protocol | UnPack | failed to ReadBody: ", err)
		return err
	}
	return nil

}

func (raw *rawProtocol) ReadHeader(data []byte, msg traffic.Message) ([]byte, error) {
	h := traffic.NewEmptyHeader()

	// 协议Id
	_ = data[0]
	// 消息序列号长度
	seqLen := data[1]
	data = data[2:]
	// 序列号内容
	ss := string(data[:seqLen])
	seqNum, err := strconv.ParseInt(ss, 36, 32)
	if err != nil {
		log.Println("raw-protocol | ReadHeader | failed to parse seq: ", err)
		return nil, err
	}
	h.SetSeq(int32(seqNum))
	data = data[seqLen:]

	// 消息类型，这个是在message里面的
	action := data[0]
	data = data[1:]
	msg.SetAction(action)

	// serviceMethod
	smLen := data[0]
	data = data[1:]

	sm := string(data[:smLen])
	h.SetServiceMethod(sm)
	data = data[smLen:]

	// body codec id
	codecId := data[0]
	data = data[1:]
	msg.SetCodec(codecId)

	msg.SetHeader(h)

	return data, nil
}

func (raw *rawProtocol) ReadBody(data []byte, msg traffic.Message) error {
	err := msg.UnMarshalBody(data)
	if err != nil {
		return err
	}
	return nil
}
