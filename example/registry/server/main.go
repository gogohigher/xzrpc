package main

import (
	"flag"
	"github.com/gogohigher/xzrpc"
	"github.com/gogohigher/xzrpc/cmd"
	"github.com/gogohigher/xzrpc/registry"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

func main() {

	// 开启服务之前，需要先开启注册中心
	go startRegistry()

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

	// 注册心跳
	var registryAddr = "http://127.0.0.1:9999/_xzrpc_/registry" // 注册中心地址
	registry.SendHeartbeat(registryAddr, "tcp@"+l.Addr().String(), 0)

	server.Accept(l)
}

// 开启注册
func startRegistry() {
	l, _ := net.Listen("tcp", "127.0.0.1:9999")

	var defaultTimeout = time.Minute * 5
	registry := registry.NewRegistry(defaultTimeout)

	var defaultPath = "/_xzrpc_/registry"
	registry.HandleHTTP(defaultPath)

	_ = http.Serve(l, nil)
}
