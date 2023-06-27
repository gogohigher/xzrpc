package main

import (
	"github.com/gogohigher/xzrpc/registry"
	"net"
	"net/http"
	"time"
)

func main() {
	startRegistry()
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
