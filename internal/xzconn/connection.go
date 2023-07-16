package xzconn

import (
	"github.com/gogohigher/xzrpc/pkg/server"
	"net"
)

type Connection struct {
	Server server.IServer // 这个Connection属于哪个Server
	// 链接ID
	ConnID  uint32
	NetConn *net.Conn
}

func NewConnection(connId uint32, netConn *net.Conn) *Connection {
	return &Connection{
		ConnID:  connId,
		NetConn: netConn,
	}
}

//func (xzConn *Connection) Start() {
//	xzConn.Server.HandleConnection(*xzConn.NetConn)
//}
