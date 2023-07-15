package codec2

import (
	"google.golang.org/protobuf/proto"
	"log"
)

var defaultProtoBufCodec = NewProtoBufCodec()

var _ Codec2 = (*ProtoBufCodec)(nil)

type ProtoBufCodec struct {
}

func NewProtoBufCodec() *ProtoBufCodec {
	return &ProtoBufCodec{}
}

func (jc *ProtoBufCodec) CodecID() byte {
	return PROTO_BUF_CODEC
}

func (jc *ProtoBufCodec) Marshal(v any) ([]byte, error) {
	if v == nil {
		log.Println("ProtoBufCodec.Marshal | param v is nil")
		return []byte{}, nil
	}
	return proto.Marshal(v.(proto.Message))
}
func (jc *ProtoBufCodec) Unmarshal(data []byte, v any) error {
	return proto.Unmarshal(data, v.(proto.Message))
}

func ProtoBufMarshal(v any) ([]byte, error) {
	return defaultProtoBufCodec.Marshal(v)
}

func ProtoBufUnmarshal(data []byte, v any) error {
	return defaultProtoBufCodec.Unmarshal(data, v)
}
