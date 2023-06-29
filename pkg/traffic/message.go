package traffic

import (
	"fmt"
	"github.com/gogohigher/xzrpc/pkg/codec2"
)

var _ Message = (*message)(nil)
var _ Header = (*header)(nil)

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
}

type message struct {
	header Header
	body   interface{}
	action byte
	codec  byte
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
	case codec2.JSON_CODEC:
		return codec2.JsonMarshal(m.body)
	default:
		return nil, fmt.Errorf("not support %d codec", m.codec)
	}
}

func (m *message) UnMarshalBody(b []byte) error {
	// TODO 暂时将body定位string类型 !!! 单元测试的时候，没有注册方法
	if m.body == nil {
		m.body = new(string)
	}
	switch m.codec {
	case codec2.JSON_CODEC:
		return codec2.JsonUnmarshal(b, m.body)
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

// Message --- end ---

// Header --- start ---
type Header interface {
	GetSeq() int32
	SetSeq(seq int32)
	GetErr() string
	GetServiceMethod() string
	SetServiceMethod(sm string)
	SetErr(err string)
}

// TODO 暂时将字段设置成可导出，因为在编解码的时候，我使用了反射
type header struct {
	ServiceMethod string // 服务名.方法名
	Seq           int32  // 序列号
	Err           string
}

func NewHeader(serviceMethod string, seq int32) Header {
	return &header{
		ServiceMethod: serviceMethod,
		Seq:           seq,
	}
}

func NewEmptyHeader() Header {
	return &header{}
}

func (h *header) GetSeq() int32 {
	return h.Seq
}

func (h *header) SetSeq(seq int32) {
	h.Seq = seq
}

func (h *header) GetErr() string {
	return h.Err
}

func (h *header) GetServiceMethod() string {
	return h.ServiceMethod
}

func (h *header) SetServiceMethod(sm string) {
	h.ServiceMethod = sm
}

func (h *header) SetErr(err string) {
	h.Err = err
}

// Header --- end ---
