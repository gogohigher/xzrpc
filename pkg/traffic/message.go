package traffic

var _ Message = (*message)(nil)
var _ Header = (*header)(nil)

// Message --- start ---
type Message interface {
}

type message struct {
	header Header
	body   interface{}
}

func NewMessage() Message {
	return &message{}
}

// Message --- end ---

// Header --- start ---
type Header interface {
	GetSeq() uint64
	GetErr() string
	GetServiceMethod() string
	SetErr(err string)
}

// TODO 暂时将字段设置成可导出，因为在编解码的时候，我使用了反射
type header struct {
	ServiceMethod string // 服务名.方法名
	Seq           uint64 // 序列号
	Err           string
}

func NewHeader(serviceMethod string, seq uint64) Header {
	return &header{
		ServiceMethod: serviceMethod,
		Seq:           seq,
	}
}

func NewEmptyHeader() Header {
	return &header{}
}

func (h *header) GetSeq() uint64 {
	return h.Seq
}

func (h *header) GetErr() string {
	return h.Err
}

func (h *header) GetServiceMethod() string {
	return h.ServiceMethod
}

func (h *header) SetErr(err string) {
	h.Err = err
}

// Header --- end ---
