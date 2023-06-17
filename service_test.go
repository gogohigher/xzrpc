package xzrpc

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

type HelloService struct {
}

type Reply string

func (hs *HelloService) Hello(hi string, reply *Reply) error {
	return errors.New("HelloService.hello")
}

func createMethodType() *methodType {
	var hs HelloService
	typ := reflect.TypeOf(&hs)
	fmt.Println(typ)

	method := typ.Method(0)
	fmt.Println(method)

	// 参数；第一个参数是hs自己
	arg0 := method.Type.In(1)
	fmt.Println(arg0)

	arg1 := method.Type.In(2)
	fmt.Println(arg1)

	// 返回参数
	return0 := method.Type.Out(0)
	fmt.Println(return0)

	return &methodType{
		method:    method,
		ArgType:   arg0,
		ReplyType: arg1,
	}
}

func TestNewArgValue(t *testing.T) {
	mt := createMethodType()
	argValue := mt.newArgValue()
	fmt.Println("argValue: ", argValue.Type())
}

func TestNewReplyValue(t *testing.T) {
	mt := createMethodType()
	replyValue := mt.newReplyValue()
	fmt.Println("replyValue: ", replyValue.Type())
}

// 测试newService

type Arith int

type Args struct {
	A, B int
}

// 满足func (t *T) MethodName(argType T1, replyType *T2) error这种格式
// 可导出的方法
func (t *Arith) Multiply(args Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

// 不可导出的方法
func (t *Arith) multiply(args Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

func TestNewService(t *testing.T) {
	var arith Arith
	s := newService(&arith)
	fmt.Println("service method len: ", len(s.method)) // 1

	mType := s.method["Multiply"]
	fmt.Println("mType is nil: ", mType == nil)
}

func TestMethodTypeCall(t *testing.T) {
	var arith Arith
	s := newService(&arith)
	mType := s.method["Multiply"]

	argValue := mType.newArgValue()
	replyValue := mType.newReplyValue()

	argValue.Set(reflect.ValueOf(Args{
		A: 2,
		B: 3,
	}))

	err := s.call(mType, argValue, replyValue)
	fmt.Println("err is nil: ", err == nil)
	fmt.Println("replyValue: ", *replyValue.Interface().(*int))

	fmt.Printf("call %d number\n", mType.numCalls)
}
