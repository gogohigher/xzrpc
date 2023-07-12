package main

import (
	"fmt"
	"github.com/gogohigher/xzrpc/apps/registry/internal/registry"
	"github.com/gogohigher/xzrpc/apps/registry/internal/server"
	"time"
)

func main() {
	// 注册
	fmt.Println(">>> xzrpc | start registry.")
	var defaultTimeout = time.Minute * 5
	registry.NewRegistry(defaultTimeout)

	s := server.NewServer()
	s.Run()
}
