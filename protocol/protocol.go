package protocol

import "github.com/gogohigher/xzrpc/traffic"

// Protocol 协议接口
type Protocol interface {
	GetContent() *Content                                                  // 协议内容
	Pack(msg traffic.Message) error                                        // 封包，即将message写入connection
	UnPack(msg traffic.Message, f func(header traffic.Header) error) error // 拆包，从connection中读取消息，读到message中
}

// Content 协议内容
type Content struct {
	Id   uint8  // 标识哪个协议
	Name string // 协议名字
}
