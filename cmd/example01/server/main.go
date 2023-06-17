package main

import (
	"github.com/gogohigher/xzrpc"
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
	server.Accept(l)
}
