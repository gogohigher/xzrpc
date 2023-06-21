package traffic

type Header struct {
	ServiceMethod string // 服务名.方法名
	Seq           uint64 // 序列号
	Error         string
}
