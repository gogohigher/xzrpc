package codec2

const (
	xxx        = iota
	JSON_CODEC = iota
)

// Codec2 新的Codec
type Codec2 interface {
	CodecID() byte // 标识当前是哪个Codec
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}
