package xzrpc

import (
	"context"
	"io"
	"log"
	"reflect"
	"sync"
)

// XZClient 支持负载均衡的client
// 1. 需要服务发现实例；2. 负载均衡策略；3. 协议选择
type XZClient struct {
	d        Discovery
	strategy StrategyMode
	option   *Option
	clients  map[string]*Client // key：rpcAddr
	mu       sync.Mutex
}

func NewXZClient(d Discovery, strategy StrategyMode, option *Option) *XZClient {
	return &XZClient{
		d:        d,
		strategy: strategy,
		option:   option,
		clients:  make(map[string]*Client),
	}
}

var _ io.Closer = (*XZClient)(nil)

// 1. 先看缓存中是否有可以的client；
// 2. 如果没有，则创建一个
// TODO 如果是拆分开加锁，会存在数据问题，如果是将整个方法都加锁，就没问题
func (xzc *XZClient) dial(rpcAddr string) (*Client, error) {
	xzc.mu.Lock()
	defer xzc.mu.Unlock()

	client, ok := xzc.clients[rpcAddr]
	if ok && !client.CheckAvailable() {
		// 存在，但不可用
		_ = client.Close()
		delete(xzc.clients, rpcAddr)
		client = nil
	}
	//xzc.mu.Unlock()

	if client == nil {
		c, err := DialRPC(rpcAddr, xzc.option)
		if err != nil {
			return nil, err
		}
		client = c
		//xzc.mu.Lock()
		xzc.clients[rpcAddr] = client
		//xzc.mu.Unlock()
	}
	return client, nil
}

// Call @xz 这块处理相当优秀
func (xzc *XZClient) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	// 1. 根据策略，选择对应的服务
	rpcAddr, err := xzc.d.Get(xzc.strategy)
	if err != nil {
		return err
	}
	log.Printf("xzclient dial rpcAddr: %s\n", rpcAddr)
	return xzc.call(ctx, rpcAddr, serviceMethod, args, reply)
}

func (xzc *XZClient) call(ctx context.Context, rpcAddr, serveMethod string, args, reply interface{}) error {
	client, err := xzc.dial(rpcAddr)
	if err != nil {
		return err
	}
	return client.Call(ctx, serveMethod, args, reply)
}

// TODO 将请求广播到所有的服务实例
func (xzc *XZClient) Broadcast(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	servers, err := xzc.d.GetAll()
	if err != nil {
		return err
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	var e error
	// 如果reply为nil，则无需赋值
	replyDone := reply == nil

	cancelContext, cancel := context.WithCancel(ctx)

	for _, rpcAddr := range servers {
		wg.Add(1)
		go func(rpcAddr string) {
			defer wg.Done()

			// @xz 为什么需要拷贝
			var copyReply interface{}
			if reply != nil {
				copyReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			err = xzc.call(cancelContext, rpcAddr, serviceMethod, args, copyReply)
			mu.Lock()
			if err != nil && e == nil {
				e = err
				cancel()
			}
			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(copyReply).Elem())
				replyDone = true
			}
			mu.Unlock()
		}(rpcAddr)
	}
	wg.Wait()
	return e
}

func (xzc *XZClient) Close() error {
	xzc.mu.Lock()

	for key, client := range xzc.clients {
		_ = client.Close()
		delete(xzc.clients, key)
	}
	xzc.mu.Unlock()
	return nil
}
