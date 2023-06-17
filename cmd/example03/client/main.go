package main

import (
	"context"
	"github.com/gogohigher/xzrpc"
	"github.com/gogohigher/xzrpc/cmd"
	"log"
	"sync"
	"time"
)

// 数据格式：option header body，由于option是json字符串，有边界的，因此这里就不会存在粘包问题
func main() {
	client, err := xzrpc.DialHTTP("tcp", "127.0.0.1:8081")
	if err != nil {
		log.Println("xzrpc client | failed to dial: ", err)
		return
	}
	defer func() {
		_ = client.Close()
	}()

	time.Sleep(time.Second)

	var wg sync.WaitGroup

	for i := 0; i < 6; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			args := &cmd.Args{
				A: i,
				B: i + 1,
			}

			var reply int
			err = client.Call(context.Background(), "Arith.Multiply", args, &reply)
			if err != nil {
				log.Fatal("xzrpc client | failed to call Arith.Multiply: ", err)
			}
			log.Printf("xzrpc client | receive reply | %d * %d = %d\n", args.A, args.B, reply)
		}(i)
	}
	wg.Wait()
}
