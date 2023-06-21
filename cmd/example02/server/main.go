package main

import (
	"github.com/gogohigher/xzrpc"
	main2 "github.com/gogohigher/xzrpc/example/registry/server"
	"log"
	"net"
)

func main() {
	l, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		log.Fatal("xzrpc server | network error:", err)
	}
	log.Println("xzrpc server | start on", l.Addr())
	server := xzrpc.NewServer()

	// 注册服务
	var arith main2.Arith
	if err = server.Register(&arith); err != nil {
		log.Fatal("xzrpc server | failed to register: ", err)
	}

	server.Accept(l)
}
