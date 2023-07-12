package gob

import (
	"fmt"
	"github.com/gogohigher/xzrpc/traffic"
	"log"
	"net"
	"testing"
)

func startServer() {
	listener, _ := net.Listen("tcp", "127.0.0.1:8082")

	for {
		conn, _ := listener.Accept()

		go handleConn(conn)
	}

}

func handleConn(conn net.Conn) {
	gob := NewGobProtocol(conn)

	for {
		receiveMsg := traffic.NewMessage()
		var reply string
		receiveMsg.SetBody(&reply)
		if err := gob.UnPack(receiveMsg, nil); err != nil {
			log.Println("gob UnPack err: ", err)
			return
		}

		log.Println("gob UnPack success reply: ", reply)
	}
}

func TestGobProtocol(t *testing.T) {
	go startServer()

	// 客户端
	conn, _ := net.Dial("tcp", "127.0.0.1:8082")
	defer func() { _ = conn.Close() }()

	gob := NewGobProtocol(conn)

	for i := 0; i < 6; i++ {

		msg := traffic.NewMessage()
		h := traffic.NewHeader("Foo.Sum", int32(i))
		msg.SetHeader(h)
		msg.SetBody(fmt.Sprintf("xzrpc req %d", h.GetSeq()))

		if err := gob.Pack(msg); err != nil {
			log.Println("gob Pack err: ", err)
			return
		}
	}

	select {}
}
