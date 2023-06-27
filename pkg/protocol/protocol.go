package protocol

import "github.com/gogohigher/xzrpc/pkg/traffic"

// Protocol 协议接口
type Protocol interface {
	// 协议内容
	GetContent() *Content
	// 封包，即将message写入connection
	Pack(traffic.Message) error
	// 拆包，从connection中读取消息，读到message中
	UnPack(traffic.Message, func(header traffic.Header) error) error
}

// Content 协议内容
type Content struct {
	Id   uint8  // 标识哪个协议
	Name string // 协议名字
}
