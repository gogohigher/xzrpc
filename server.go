package xzrpc

import (
	"errors"
	"fmt"
	"github.com/gogohigher/xzrpc/codec"
	"github.com/gogohigher/xzrpc/pkg/codec2"
	_const "github.com/gogohigher/xzrpc/pkg/const"
	"github.com/gogohigher/xzrpc/pkg/protocol"
	"github.com/gogohigher/xzrpc/pkg/traffic"
	"github.com/gogohigher/xzrpc/protocol/raw"
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
}

func NewServer() *Server {
	return &Server{}
}

// TODO-这里可以优化的，看下xzlink
func (s *Server) Accept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("failed to listener accept: ", err)
			continue
		}
		// TODO 要不要做一些前置检查
		go s.HandleConn(conn)
	}
}

// 首先使用 json.NewDecoder 反序列化得到 Option 实例，检查 MagicNumber 和 CodeType 的值是否正确。
// 然后根据 CodeType 得到对应的消息编解码器，接下来的处理交给 serverCodec。
func (s *Server) HandleConn(conn io.ReadWriteCloser) {
	//var option Option
	//err := json.NewDecoder(conn).Decode(&option)
	//if err != nil {
	//	log.Println("failed to decode json: ", err)
	//	return
	//}
	//log.Printf("HandleConn option: %+v\n", option)

	//if option.Magic != Magic {
	//	log.Printf("valid magic number: %v\n", option.Magic)
	//	return
	//}

	//codecFunc, ok := codec.NewCodecFuncMap[option.CodecType]
	//if !ok {
	//	log.Printf("failed to find %v CodecType\n", option.CodecType)
	//	return
	//}

	rawProtocol := raw.NewRawProtocol(conn)
	s.HandleCodec(nil, rawProtocol)
}

var EmptyData = struct{}{}

// 根据编解码器处理数据
// 并发处理请求，有序返回数据
func (s *Server) HandleCodec(cc codec.Codec, rawProtocol protocol.Protocol) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		req, err := s.readRequest(cc, rawProtocol)
		if err != nil {
			if req == nil {
				break
			}

			req.message.Header().SetErr(err.Error())
			s.sendResp(cc, req.message.Header(), EmptyData, sending, rawProtocol)
			continue
		}
		// 并发处理
		wg.Add(1)
		go s.handleRequest(cc, req, sending, wg, rawProtocol)
	}
	wg.Wait()
	//_ = cc.Close()
}

type request struct {
	message traffic.Message
	//header      traffic.Header
	args, reply reflect.Value

	srv   *service
	mType *methodType
}

func (s *Server) readRequest(cc codec.Codec, rawProtocol protocol.Protocol) (*request, error) {
	// 1. read header
	//var h = traffic.NewEmptyHeader()

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

	//if err := cc.ReadHeader(h); err != nil {
	//	fmt.Println("failed to ReadHeader: ", err)
	//	return nil, err
	//}
	//req := &request{header: h}
	////req.args = reflect.New(reflect.TypeOf(""))
	//srv, mType, err := s.findService(h.GetServiceMethod())
	//if err != nil {
	//	log.Println("xzrpc server | readRequest | failed to findService: ", err)
	//	return req, err
	//}
	//req.srv = srv
	//req.mType = mType
	//
	//req.args = req.mType.newArgValue()
	//req.reply = req.mType.newReplyValue()
	//
	//// @xz req.args可能是值类型，也可能是指针类型
	//args := req.args.Interface()
	//if req.args.Kind() != reflect.Pointer {
	//	args = req.args.Addr().Interface()
	//}
	//
	//// 2. read body，ReadBody的参数必须是指针类型，因此上述需要对req.args做判断
	//if err := cc.ReadBody(args); err != nil {
	//	log.Println("xzrpc server | readRequest | failed to ReadBody: ", err)
	//	return req, err
	//}
	//// 3. header成功，但是body失败了，也返回
	//return req, nil
}

// TODO wg是不是可以绑定到Server中，作为它的属性
// TODO 服务端处理超时
func (s *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup, rawProtocol protocol.Protocol) {
	defer wg.Done()

	//log.Println("server.handleRequest | ", req.header, req.args.Elem())
	//req.reply = reflect.ValueOf(fmt.Sprintf("xzrpc resp %d", req.header.Seq))

	// 调用真正方法
	err := req.srv.call(req.mType, req.args, req.reply)
	if err != nil {
		req.message.Header().SetErr(err.Error())

		// 发送一个空结构体
		s.sendResp(cc, req.message.Header(), struct{}{}, sending, rawProtocol)
		return
	}

	s.sendResp(cc, req.message.Header(), req.reply.Interface(), sending, rawProtocol)

}

func (s *Server) sendResp(cc codec.Codec, header traffic.Header, body interface{}, sending *sync.Mutex, rawProtocol protocol.Protocol) {
	// 有序发送
	sending.Lock()
	defer sending.Unlock()

	// 改成raw-protocol
	//if err := cc.Write(header, body); err != nil {
	//	log.Println("failed to write resp: ", err)
	//}

	msg := traffic.NewMessage()
	msg.SetHeader(header)
	msg.SetBody(body)
	msg.SetAction(traffic.CALL) // TODO 这里要改成REPLY吗？需要定义一下CALL、REPLY等等
	msg.SetCodec(codec2.JSON_CODEC)

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
	s.HandleConn(conn)
}

func (s *Server) HandleHTTP() {
	// 默认rpc path
	http.Handle(_const.DEFAULT_XZRPC_PATH, s)
}
