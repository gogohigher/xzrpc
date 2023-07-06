package traffic

import (
	"fmt"
	codec22 "github.com/gogohigher/xzrpc/codec2"
	"sync"
)

var (
	_           Message = (*message)(nil)
	_           Header  = (*header)(nil)
	MessagePool sync.Pool
)

// Message --- start ---
type Message interface {
	GetSeq() int32
	Action() byte // 消息类型
	SetAction(action byte)
	ServiceMethod() string
	Codec() byte // 针对Body
	SetCodec(c byte)
	MarshalBody() ([]byte, error)
	UnMarshalBody(b []byte) error
	SetHeader(h Header)
	Header() Header
	SetBody(b interface{})
	Body() any
	SetCompressor(compressType byte)
	Compressor() byte
	ResetMessage()
}

type message struct {
	header       Header
	body         interface{}
	action       byte
	codec        byte
	compressType byte // 压缩算法
}

func init() {
	MessagePool = sync.Pool{New: func() any {
		return NewMessage()
	}}
}

func NewMessage() Message {
	return &message{}
}

func (m *message) GetSeq() int32 {
	return m.header.GetSeq()
}

func (m *message) Action() byte {
	return m.action
}

func (m *message) SetAction(action byte) {
	m.action = action
}

func (m *message) ServiceMethod() string {
	return m.header.GetServiceMethod()
}

func (m *message) Codec() byte {
	return m.codec
}

func (m *message) SetCodec(c byte) {
	m.codec = c
}

func (m *message) MarshalBody() ([]byte, error) {
	switch m.codec {
	case codec22.JSON_CODEC:
		return codec22.JsonMarshal(m.body)
	default:
		return nil, fmt.Errorf("not support %d codec", m.codec)
	}
}

func (m *message) UnMarshalBody(b []byte) error {
	// TODO 暂时将body定位string类型 !!! 单元测试的时候，没有注册方法
	//if m.body == nil {
	//	m.body = new(string)
	//}
	switch m.codec {
	case codec22.JSON_CODEC:
		return codec22.JsonUnmarshal(b, m.body)
	default:
		return fmt.Errorf("not support %d codec", m.codec)
	}
}

func (m *message) SetHeader(h Header) {
	m.header = h
}

func (m *message) Header() Header {
	return m.header
}

func (m *message) SetBody(b interface{}) {
	m.body = b
}

func (m *message) Body() any {
	return m.body
}

func (m *message) SetCompressor(compressType byte) {
	m.compressType = compressType
}

func (m *message) Compressor() byte {
	return m.compressType
}

func (m *message) ResetMessage() {
	m.header = nil
	m.body = nil
	m.action = 0
	m.codec = 0
	m.compressType = 0
}
