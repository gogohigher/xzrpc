package xzrpc

import (
	"github.com/gogohigher/xzrpc/codec"
	"time"
)

const Magic = 0x9f9f

type Option struct {
	Magic     int
	CodecType codec.Type
	// 链接超时，@xz暂时先放在这里，后面优化，拿走
	ConnectTimeOut time.Duration
	// 处理超时
	HandleTimeOut time.Duration
}

var DefaultOption = Option{
	Magic:          Magic,
	CodecType:      codec.GobType,
	ConnectTimeOut: 0,
}
