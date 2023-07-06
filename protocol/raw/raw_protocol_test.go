package raw

import (
	"github.com/gogohigher/xzrpc/codec2"
	"github.com/gogohigher/xzrpc/traffic"
	"log"
	"net"
	"testing"
	"time"
)

func TestPack(t *testing.T) {
	go startServer()

	conn, err := net.Dial("tcp", "127.0.0.1:8081")
	if err != nil {
		log.Println("err: ", err)
		return
	}
	defer func() { _ = conn.Close() }()
	rawProtocol := NewRawProtocol(conn)

	msg := createMessage()

	err = rawProtocol.Pack(msg)
	if err != nil {
		log.Println("TestPack Pack err: ", err)
		return
	}

	time.Sleep(time.Hour)
}

func createMessage() traffic.Message {
	msg := traffic.NewMessage()
	msg.SetAction(traffic.CALL)
	msg.SetCodec(codec2.JSON_CODEC)
	msg.SetBody("this is raw protocol")

	// header
	h := traffic.NewHeader("HelloService.Hello", 1)
	msg.SetHeader(h)

	return msg
}

func startServer() {
	l, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		log.Fatal("xzrpc server | network error:", err)
	}
	log.Println("xzrpc server | start on", l.Addr())
	conn, _ := l.Accept()

	// 下面是测试UnPack
	rawProtocol := NewRawProtocol(conn)

	msg := traffic.NewMessage()
	err = rawProtocol.UnPack(msg, nil)
	if err != nil {
		log.Println("TestPack UnPack err: ", err)
		return
	}
	log.Printf("TestPack UnPack msg: %+v\n", msg)
}
