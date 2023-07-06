package main

import (
	"encoding/json"
	"fmt"
	"github.com/gogohigher/xzrpc"
	"github.com/gogohigher/xzrpc/codec"
	"github.com/gogohigher/xzrpc/traffic"
	"log"
	"net"
	"time"
)

// 数据格式：option header body，由于option是json字符串，有边界的，因此这里就不会存在粘包问题
func main() {
	conn, _ := net.Dial("tcp", "127.0.0.1:8081")
	defer func() { _ = conn.Close() }()

	time.Sleep(time.Second)

	// send options
	_ = json.NewEncoder(conn).Encode(xzrpc.DefaultOption)

	cc := codec.NewGobCodec(conn)

	for i := 0; i < 6; i++ {
		//h := &traffic.Header{
		//	ServiceMethod: "Foo.Sum",
		//	Seq:           uint64(i),
		//}
		h := traffic.NewHeader("Foo.Sum", int32(i))
		// 写数据到服务端
		_ = cc.Write(h, fmt.Sprintf("xzrpc req %d", h.GetSeq()))

		// 下面是返回数据
		_ = cc.ReadHeader(h)
		var reply string
		_ = cc.ReadBody(&reply)
		log.Println("reply:", reply)
	}
}
