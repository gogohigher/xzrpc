package main

import (
	"context"
	"github.com/gogohigher/xzrpc"
	"github.com/gogohigher/xzrpc/cmd"
	"log"
	"sync"
	"time"
)

func main() {
	registryAddr := "http://127.0.0.1:9999/_xzrpc_/registry"
	call(registryAddr)
}

func call(registry string) {
	d := xzrpc.NewRegistryDiscovery(registry, 0)

	client := xzrpc.NewXZClient(d, xzrpc.RandomStrategy, nil)
	defer func() {
		_ = client.Close()
	}()

	var wg sync.WaitGroup
	for i := 1; i < 6; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &cmd.Args{
				A: i,
				B: i + 1,
			}
			arith(context.Background(), client, "call", "Arith.Multiply", args, i)
		}(i)
	}
	wg.Wait()
}

func broadcast(registry string) {
	d := xzrpc.NewRegistryDiscovery(registry, 0)

	client := xzrpc.NewXZClient(d, xzrpc.RandomStrategy, nil)
	defer func() {
		_ = client.Close()
	}()

	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &cmd.Args{
				A: i,
				B: i + 1,
			}
			arith(context.Background(), client, "broadcast", "Arith.Multiply", args, i)

			ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
			arith(ctx, client, "broadcast", "Arith.MultiplySleep", args, i)

		}(i)
	}
	wg.Wait()
}

func arith(ctx context.Context, xzc *xzrpc.XZClient, action, serviceMethod string, args *cmd.Args, i int) {
	var reply int
	var err error

	switch action {
	case "call":
		err = xzc.Call(ctx, serviceMethod, args, &reply)
	case "broadcast":
		err = xzc.Broadcast(ctx, serviceMethod, args, &reply)
	}
	if err != nil {
		log.Printf("[No.%d] %s %s error: %v", i, action, serviceMethod, err)
	} else {
		log.Printf("[No.%d] %s %s success: %d * %d = %d", i, action, serviceMethod, args.A, args.B, reply)
	}
}
