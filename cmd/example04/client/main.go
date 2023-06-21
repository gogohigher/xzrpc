package main

import (
	"context"
	"github.com/gogohigher/xzrpc"
	main2 "github.com/gogohigher/xzrpc/example/registry/server"
	"log"
	"sync"
	"time"
)

func call(addr1, addr2 string) {
	d := xzrpc.NewDefaultDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	client := xzrpc.NewXZClient(d, xzrpc.RandomStrategy, nil)
	defer func() {
		_ = client.Close()
	}()

	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &main2.Args{
				A: i,
				B: i + 1,
			}
			arith(context.Background(), client, "call", "Arith.Multiply", args)
		}(i)
	}
	wg.Wait()
}

func broadcast(addr1, addr2 string) {
	d := xzrpc.NewDefaultDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	client := xzrpc.NewXZClient(d, xzrpc.RandomStrategy, nil)
	defer func() {
		_ = client.Close()
	}()

	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &main2.Args{
				A: i,
				B: i + 1,
			}
			arith(context.Background(), client, "broadcast", "Arith.Multiply", args)

			ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
			arith(ctx, client, "broadcast", "Arith.MultiplySleep", args)

		}(i)
	}
	wg.Wait()
}

func arith(ctx context.Context, xzc *xzrpc.XZClient, action, serviceMethod string, args *main2.Args) {
	var reply int
	var err error

	switch action {
	case "call":
		err = xzc.Call(ctx, serviceMethod, args, &reply)
	case "broadcast":
		err = xzc.Broadcast(ctx, serviceMethod, args, &reply)
	}
	if err != nil {
		log.Printf("%s %s error: %v", action, serviceMethod, err)
	} else {
		log.Printf("%s %s success: %d * %d = %d", action, serviceMethod, args.A, args.B, reply)
	}

}

// TODO 负载均衡测试

func main() {
	addr1 := "127.0.0.1:8081"
	addr2 := "127.0.0.1:8082"
	// TODO 丢数据了，循环6次，但是只打印了5次
	call(addr1, addr2)

	// TODO 同样存在问题，而且还遇到死锁
	//broadcast(addr1, addr2)
}
