package codec2

import (
	"encoding/json"
)

var defaultJsonCodec = NewJsonCodec()

type JsonCodec struct {
}

func NewJsonCodec() *JsonCodec {
	return &JsonCodec{}
}

func (jc *JsonCodec) CodecID() byte {
	return JSON_CODEC
}

func (jc *JsonCodec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}
func (jc *JsonCodec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func JsonMarshal(v any) ([]byte, error) {
	return defaultJsonCodec.Marshal(v)
}

func JsonUnmarshal(data []byte, v any) error {
	return defaultJsonCodec.Unmarshal(data, v)
}
