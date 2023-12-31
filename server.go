package xzrpc

import (
	"errors"
	"fmt"
	"github.com/gogohigher/xzrpc/internal/const"
	"github.com/gogohigher/xzrpc/internal/xzconn"
	"github.com/gogohigher/xzrpc/pkg/pool"
	"github.com/gogohigher/xzrpc/protocol"
	"github.com/gogohigher/xzrpc/protocol/raw"
	"github.com/gogohigher/xzrpc/traffic"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

//首先定义了结构体 Server，没有任何的成员字段。
//实现了 Accept 方式，net.Listener 作为参数，for 循环等待 socket 连接建立，并开启子协程处理，处理过程交给了 ServerConn 方法。
//DefaultServer 是一个默认的 Server 实例，主要为了用户使用方便。

type Server struct {
	serviceMap sync.Map
	workerPool *pool.WorkerPool
}

func NewServer() *Server {
	s := &Server{
		workerPool: pool.NewDefaultWorkerPool(),
	}
	// @xz 是否应该放在这里？再考虑一下
	s.Start()
	return s
}

func (s *Server) Start() {
	s.workerPool.Start()
}

// Accept TODO-这里可以优化的，看下xzlink
// TODO 连接这套也上了
func (s *Server) Accept(listener net.Listener) {
	var connID uint32 = 0

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("failed to listener accept: ", err)
			continue
		}
		xzConn := xzconn.NewConnection(connID, &conn)
		connID++
		s.HandleConnection(xzConn)
	}
}

// HandleConnection 首先使用 json.NewDecoder 反序列化得到 Option 实例，检查 MagicNumber 和 CodeType 的值是否正确。
// 然后根据 CodeType 得到对应的消息编解码器，接下来的处理交给 serverCodec。
func (s *Server) HandleConnection(xzConn *xzconn.Connection) {
	workerId := xzConn.ConnID % uint32(s.workerPool.PoolSize)
	s.workerPool.Workers[workerId].Enqueue(func() {
		rawProtocol := raw.NewRawProtocol(*xzConn.NetConn)
		s.HandleCodec(rawProtocol)
	})
}

var EmptyData = struct{}{}

// HandleCodec 根据编解码器处理数据
// 并发处理请求，有序返回数据
func (s *Server) HandleCodec(rawProtocol protocol.Protocol) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		req, err := s.readRequest(rawProtocol)
		if err != nil {
			if req == nil {
				break
			}

			req.message.Header().SetErr(err.Error())
			s.sendResp(req.message, EmptyData, sending, rawProtocol)
			continue
		}
		// 并发处理
		wg.Add(1)
		go s.handleRequest(req, sending, wg, rawProtocol)
	}
	wg.Wait()
}

type request struct {
	message     traffic.Message
	args, reply reflect.Value

	srv   *service
	mType *methodType
}

func (s *Server) readRequest(rawProtocol protocol.Protocol) (*request, error) {
	// 1. read header
	req := &request{}

	msg := traffic.NewMessage()
	req.message = msg

	err := rawProtocol.UnPack(msg, func(header traffic.Header) error {
		srv, mType, err := s.findService(header.GetServiceMethod())
		if err != nil {
			log.Println("xzrpc server | readRequest | failed to findService: ", err)
			return err
		}
		req.srv = srv
		req.mType = mType
		req.args = req.mType.newArgValue()
		req.reply = req.mType.newReplyValue()
		args := req.args.Interface()
		if req.args.Kind() != reflect.Pointer {
			args = req.args.Addr().Interface()
		}
		msg.SetBody(args)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return req, err
}

// TODO wg是不是可以绑定到Server中，作为它的属性
// TODO 服务端处理超时
func (s *Server) handleRequest(req *request, sending *sync.Mutex, wg *sync.WaitGroup, rawProtocol protocol.Protocol) {
	defer wg.Done()

	//log.Println("server.handleRequest | ", req.header, req.args.Elem())
	//req.reply = reflect.ValueOf(fmt.Sprintf("xzrpc resp %d", req.header.Seq))

	// 调用真正方法
	err := req.srv.call(req.mType, req.args, req.reply)
	if err != nil {
		req.message.Header().SetErr(err.Error())

		// 发送一个空结构体
		s.sendResp(req.message, struct{}{}, sending, rawProtocol)
		return
	}

	s.sendResp(req.message, req.reply.Interface(), sending, rawProtocol)

}

func (s *Server) sendResp(in traffic.Message, body interface{}, sending *sync.Mutex, rawProtocol protocol.Protocol) {
	// 有序发送
	sending.Lock()
	defer sending.Unlock()

	msg := traffic.MessagePool.Get().(traffic.Message)
	defer func() {
		msg.ResetMessage()
		traffic.MessagePool.Put(msg)
	}()

	msg.SetHeader(in.Header())
	msg.SetBody(body)
	msg.SetAction(in.Action()) // TODO 这里要改成REPLY吗？需要定义一下CALL、REPLY等等
	msg.SetCodec(in.Codec())
	msg.SetCompressor(in.Compressor())
	log.Println(">>> 服务端 in.Action: ", in.Action(), ", in.Codec: ", in.Codec(), ", in.Compressor: ", in.Compressor())

	if err := rawProtocol.Pack(msg); err != nil {
		log.Println("failed to write resp: ", err)
	}
}

func (s *Server) Register(instance interface{}) error {
	srv := newService(instance)

	_, loaded := s.serviceMap.LoadOrStore(srv.name, srv)
	// The loaded result is true if the value was loaded, false if stored.
	if loaded {
		// 重复添加
		return fmt.Errorf("xzrpc server | %s has already exist", srv.name)
	}
	return nil
}

// 通过serviceMethod，从Server的serviceMap中找到对应的service
func (s *Server) findService(serviceMethod string) (srv *service, mType *methodType, err error) {

	// 1. 根据服务名，找到对应的service
	// 2. 根据service，就能找到对应的methodType

	dotIndex := strings.LastIndex(serviceMethod, ".")
	if dotIndex < 0 {
		return nil, nil, errors.New("xzrpc Server | findService | failed to parse serviceMethod")
	}

	serviceName := serviceMethod[:dotIndex]
	methodName := serviceMethod[dotIndex+1:]
	value, ok := s.serviceMap.Load(serviceName)
	if !ok {
		return nil, nil, fmt.Errorf("xzrpc Server | findService | failed to find %s service", serviceName)
	}

	srv = value.(*service)
	mType = srv.method[methodName]
	if mType == nil {
		return nil, nil, fmt.Errorf("xzrpc Server | findService | failed to find %s method", methodName)
	}
	return
}

// 实现Handler接口
// http.Handle("xx", handler)
// 这里的HTTP请求只允许处理CONNECT的请求
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain;charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = io.WriteString(w, "only support CONNECT  method\n")
		return
	}

	// Server接管connection
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Printf("rpc server | failed to  to hijacking %s : %s\n", r.RemoteAddr, err.Error())
		return
	}
	_, _ = io.WriteString(conn, "HTTP/1.0 "+_const.CONNECTED+"\n\n")

	xzConn := xzconn.NewConnection(0, &conn)
	s.HandleConnection(xzConn)
}

func (s *Server) HandleHTTP() {
	// 默认rpc path
	http.Handle(_const.DEFAULT_XZRPC_PATH, s)
}
