package main

import (
	"github.com/gogohigher/xzrpc"
	"github.com/gogohigher/xzrpc/cmd"
	"log"
	"net"
	"net/http"
)

func main() {
	l, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		log.Fatal("xzrpc server | network error:", err)
	}
	log.Println("xzrpc server | start on", l.Addr())
	server := xzrpc.NewServer()

	// 注册服务
	var arith cmd.Arith
	if err = server.Register(&arith); err != nil {
		log.Fatal("xzrpc server | failed to register: ", err)
	}

	// listener.Accept()改成http.Handle
	//server.Accept(l)
	server.HandleHTTP()
	log.Fatal(http.Serve(l, nil))
}
