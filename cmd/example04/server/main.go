package main

import (
	"flag"
	"github.com/gogohigher/xzrpc"
	"github.com/gogohigher/xzrpc/cmd"
	"log"
	"net"
	"strconv"
)

func main() {
	portPtr := flag.Int("port", 8081, "the port number")
	flag.Parse()
	port := *portPtr
	log.Println("port: ", port)

	l, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
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

	server.Accept(l)
}
