package traffic

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
