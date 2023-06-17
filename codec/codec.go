package codec

import "io"

type Header struct {
	ServiceMethod string // 服务名.方法名
	Seq           uint64 // 序列号
	Error         string
}

// Codec 编解码器
type Codec interface {
	io.Closer
	ReadHeader(header *Header) error
	ReadBody(body interface{}) error
	Write(header *Header, body interface{}) error
}

type Type string

const (
	GobType  = "application/gob"
	JsonType = "application/json"
)

type NewCodecFunc func(closer io.ReadWriteCloser) Codec

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
