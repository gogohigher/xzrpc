package xzrpc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/gogohigher/xzrpc/pkg/codec2"
	_const "github.com/gogohigher/xzrpc/pkg/const"
	"github.com/gogohigher/xzrpc/pkg/traffic"
	"github.com/gogohigher/xzrpc/protocol/raw"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

var _ io.Closer = (*Client)(nil)

// 看下net/rpc，简直一模一样
type Call struct {
	Seq           uint64
	ServiceMethod string // 服务名.方法名
	Args          interface{}
	Reply         interface{}
	Error         error
	Done          chan *Call
}

// Client xzrpc client
type Client struct {
	//cc     codec.Codec // @xz 这个可以删除
	option *Option
	//header    traffic.Header
	seq       uint64
	taskQueue map[uint64]*Call
	closed    bool
	shutdown  bool // @xz 和close含义有点区别
	conn      io.ReadWriteCloser

	sendMu sync.Mutex
	mu     sync.Mutex
}

type createClientResult struct {
	client *Client
	err    error
}

type newClientFunc func(conn net.Conn, option *Option) (*Client, error)

// NewClient 创建Client实例
// 1. 协议交换，即发生Option信息给服务端
// 2. 创建goroutine调用receive接收响应
func NewClient(conn net.Conn, option *Option) (*Client, error) {
	// 1. 根据codecType，找到对应的CodeC构造器
	//f, ok := codec.NewCodecFuncMap[option.CodecType]
	//if !ok {
	//	_ = conn.Close()
	//	return nil, fmt.Errorf("%s is invalid codec type\n", option.CodecType)
	//}

	// 不再需要提前发，将编解码协议都写到header中
	// 2. 发送协议给服务端
	//err := json.NewEncoder(conn).Encode(option)
	//if err != nil {
	//	_ = conn.Close()
	//	return nil, err
	//}

	// 3. 创建客户端
	client := &Client{
		seq: 1, // TODO 暂时写死
		//cc:        f(conn),
		option:    option,
		taskQueue: make(map[uint64]*Call),
		conn:      conn,
	}

	// 4. 接口响应
	go client.receive()
	return client, nil
}

// NewHTTPClient 使用HTTP创建一个client实例
// CONNECT localhost:8081/_xzrpc_ http/1.0
// HTTP/1.0 200 Connected to xzrpc
// 通过HTTP CONNECT请求建立连接之后，后续的通信都交给xzrpc.Client
func NewHTTPClient(conn net.Conn, option *Option) (*Client, error) {
	_, _ = io.WriteString(conn, fmt.Sprintf("CONNECT %s HTTP/1.0\n\n", _const.DEFAULT_XZRPC_PATH))

	request := http.Request{Method: "CONNECT"}
	resp, err := http.ReadResponse(bufio.NewReader(conn), &request)
	if err == nil {
		if resp.Status == _const.CONNECTED {
			return NewClient(conn, option)
		}
		return nil, errors.New("xzrpc client | unexpected HTTP resp: " + resp.Status)
	}
	return nil, err
}

// Dial 拨号，链接服务端
func Dial(network, address string, opts ...*Option) (client *Client, err error) {
	f := NewClient
	return dialTimeOut(f, network, address, opts...)
}

// DialHTTP 先建立HTTP，然后进行RPC通信
func DialHTTP(network, address string, opts ...*Option) (client *Client, err error) {
	f := NewHTTPClient
	return dialTimeOut(f, network, address, opts...)
}

// DialRPC rpcAddr：格式为 protocol@addr，例如：tcp@127.0.0.1:8081、http@127.0.0.1:8081
func DialRPC(rpcAddr string, opts ...*Option) (*Client, error) {
	parts := strings.Split(rpcAddr, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("xzrp client | faield to parse rpcAddr: %s", rpcAddr)
	}

	protocol := parts[0]
	addr := parts[1]

	switch protocol {
	case "http":
		return DialHTTP(protocol, addr, opts...)
	case "tcp":
		return Dial(protocol, addr, opts...)
	default:
		return nil, fmt.Errorf("xzrp client | not support %s protocol", protocol)
	}
}

// 超时：包括 dial + create client
// 将NewClient操作提到入参，作为函数参数，这样也方便单元测试
func dialTimeOut(f newClientFunc, network, address string, opts ...*Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTimeout(network, address, opt.ConnectTimeOut)
	if err != nil {
		return nil, err
	}

	// 如果返回的client有错误，应该关闭conn
	defer func() {
		if client == nil || err != nil {
			_ = conn.Close()
		}
	}()

	ch := make(chan createClientResult)

	go func() {
		client, err = f(conn, opt)
		ch <- createClientResult{
			client: client,
			err:    err,
		}
	}()

	// 不限制时间，就一直阻塞住
	if opt.ConnectTimeOut <= 0 {
		res := <-ch
		return res.client, res.err
	}

	// 超时逻辑
	select {
	case <-time.After(opt.ConnectTimeOut):
		return nil, fmt.Errorf("xzrpc client | connect timeout")
	case res := <-ch:
		return res.client, res.err
	}

}

// 接收服务端的响应
// 1. call不存在
// 2. call存在，但服务端处理出错
// 3. call存在，服务端处理正常
func (c *Client) receive() {
	var err error
	var call *Call
	for err == nil {
		//var h = traffic.NewEmptyHeader()
		//err = c.cc.ReadHeader(h)

		msg := traffic.NewMessage()
		rawProtocol := raw.NewRawProtocol(c.conn)
		err = rawProtocol.UnPack(msg, func(header traffic.Header) error {

			call = c.removeCall(uint64(header.GetSeq()))
			switch {
			case call == nil:
				// @xz 没看懂
				msg.SetBody(nil)
			case header.GetErr() != "":
				call.Error = fmt.Errorf(header.GetErr())
				msg.SetBody(nil)
				call.done()

			default:
				msg.SetBody(call.Reply) // 设置一下，其实可以理解成初始化
				//err = c.cc.ReadBody(call.Reply)
				//if err != nil {
				//	call.Error = errors.New("read body " + err.Error())
				//}
				//call.done()
			}
			return nil
		})

		if call != nil {
			call.done()
		}

		//if err != nil {
		//	break
		//}
		// call存在
		//call := c.removeCall(uint64(h.GetSeq()))
		//switch {
		//case call == nil:
		//	// @xz 没看懂
		//	err = c.cc.ReadBody(nil)
		//case h.GetErr() != "":
		//	call.Error = fmt.Errorf(h.GetErr())
		//	err = c.cc.ReadBody(nil)
		//	call.done()
		//
		//default:
		//	err = c.cc.ReadBody(call.Reply)
		//	if err != nil {
		//		call.Error = errors.New("read body " + err.Error())
		//	}
		//	call.done()
		//}
	}
	c.terminateCalls(err)
}

// 发送请求
func (c *Client) send(call *Call) {
	c.sendMu.Lock() // @xz 发送请求为啥要加锁
	defer c.sendMu.Unlock()

	seq, err := c.registerCall(call)
	if err != nil {
		call.Error = err
		// 一次调用结束
		call.done()
		return
	}

	// prepare req header
	//c.header.ServiceMethod = call.ServiceMethod
	//c.header.Seq = seq
	//c.header.Error = ""

	//c.header = traffic.NewHeader(call.ServiceMethod, int32(seq))

	h := traffic.NewHeader(call.ServiceMethod, int32(seq))

	rawProtocol := raw.NewRawProtocol(c.conn)
	// 构造消息
	msg := traffic.NewMessage()
	msg.SetHeader(h)
	msg.SetBody(call.Args)
	msg.SetAction(traffic.CALL)
	msg.SetCodec(codec2.JSON_CODEC)

	err = rawProtocol.Pack(msg)

	// send req with encode
	//err = c.cc.Write(c.header, call.Args)

	if err != nil {
		_ = c.removeCall(seq)
		// @xz removeCall会为空吗？
	}
}

// Go 异步调用
// 返回一个call对象，这个call对象表示一次调用
func (c *Client) Go(serviceMethod string, args, reply interface{}, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10) // @xz 10，参考了其他文章；为啥要有缓冲
	}
	if cap(done) == 0 {
		log.Panic("rpc client | done channel must have buffer")
	}

	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}

	c.send(call)
	return call
}

// Call 同步调用
// ctx超时处理
func (c *Client) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	call := c.Go(serviceMethod, args, reply, make(chan *Call, 1)) // channel的容量是1，阻塞住call.Done，等待响应返回，只要有一个call来了，就返回了

	select {
	case <-ctx.Done():
		// 超时
		c.removeCall(call.Seq)
		return fmt.Errorf("xzrpc client | call time out: %v", ctx.Err().Error())
	case call := <-call.Done:
		return call.Error
	}
}

// 将参数call添加到任务队列，更新seq
func (c *Client) registerCall(call *Call) (uint64, error) {
	if c.closed || c.shutdown {
		return 0, errors.New("connection has closed")
	}
	c.mu.Lock()
	call.Seq = c.seq
	c.taskQueue[call.Seq] = call
	c.seq++
	c.mu.Unlock()
	return call.Seq, nil
}

// 根据seq，删除对应的call
func (c *Client) removeCall(seq uint64) *Call {
	c.mu.Lock()
	defer c.mu.Unlock()
	call, ok := c.taskQueue[seq]
	if !ok {
		return nil
	}
	delete(c.taskQueue, seq)
	return call
}

// 服务端或客户端发送错误时调用，将shutdown设置为true，且将错误信息通知任务队列中的call
func (c *Client) terminateCalls(err error) {
	// @xz 有必要使用两个锁吗?

	c.sendMu.Lock()
	defer c.sendMu.Unlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.shutdown = true
	for _, call := range c.taskQueue {
		call.Error = err
		call.done()
	}
}

// Close 关闭connection
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return errors.New("connection has closed")
	}
	c.closed = true
	//return c.cc.Close()
	return nil
}

func (c *Client) CheckAvailable() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return !c.closed && !c.shutdown
}

func (call *Call) done() {
	call.Done <- call
}

func parseOptions(opts ...*Option) (*Option, error) {
	// 1. 判断是否为空
	if len(opts) == 0 || opts[0] == nil {
		return &DefaultOption, nil
	}
	// 2. 判断是否超过一个
	if len(opts) > 1 {
		return nil, errors.New("options must not be more than one")
	}
	return opts[0], nil
}
