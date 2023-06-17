package xzrpc

import (
	"log"
	"net"
	"testing"
	"time"
)

func TestDialTimeOut(t *testing.T) {
	t.Parallel()
	l, _ := net.Listen("tcp", "127.0.0.1:8081")

	f := func(conn net.Conn, opt *Option) (client *Client, err error) {
		_ = conn.Close()
		time.Sleep(time.Second * 2)
		return nil, nil
	}
	t.Run("timeout", func(t *testing.T) {
		_, err := dialTimeOut(f, "tcp", l.Addr().String(), &Option{ConnectTimeOut: time.Second})
		log.Println("case1: ", err)
	})
	t.Run("0", func(t *testing.T) {
		_, err := dialTimeOut(f, "tcp", l.Addr().String(), &Option{ConnectTimeOut: 0})
		log.Println("case2: ", err)
	})
}
