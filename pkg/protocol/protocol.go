package protocol

import "github.com/gogohigher/xzrpc/pkg/traffic"

// Protocol 协议接口
type Protocol interface {
	// GetContent 协议内容
	GetContent() *Content
	// Pack 封包，即将message写入connection
	Pack(traffic.Message) error
	// UnPack 拆包，从connection中读取消息，读到message中
	UnPack(traffic.Message) error
}

// Content 协议内容
type Content struct {
	Id   uint8  // 标识哪个协议
	Name string // 协议名字
}
