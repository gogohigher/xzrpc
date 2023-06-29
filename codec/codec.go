package codec

import (
	"github.com/gogohigher/xzrpc/pkg/traffic"
	"io"
)

// Codec 编解码器
type Codec interface {
	io.Closer
	ReadHeader(header traffic.Header) error
	ReadBody(body interface{}) error
	Write(header traffic.Header, body interface{}) error
}

type Type string

const (
	GobType  = "application/gob"
	JsonType = "application/json"
)

type NewCodecFunc func(closer io.ReadWriteCloser) Codec

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	//NewCodecFuncMap = make(map[Type]NewCodecFunc)
	//NewCodecFuncMap[GobType] = NewGobCodec
}
