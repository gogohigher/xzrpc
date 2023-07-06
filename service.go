package xzrpc

import (
	"github.com/gogohigher/xzrpc/internal/utils"
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

type methodType struct {
	// 方法本身
	method reflect.Method
	// 第一个参数的类型
	ArgType reflect.Type
	// 第二个参数的类型
	ReplyType reflect.Type
	// 后续统计方法调用次数时会用到
	numCalls uint64
}

// 创建实例
func (m *methodType) newArgValue() reflect.Value {
	var argValue reflect.Value
	if m.ArgType.Kind() == reflect.Pointer {
		// It panics if the type's Kind is not Array, Chan, Map, Pointer, or Slice.
		argValue = reflect.New(m.ArgType.Elem())
	} else {
		// It panics if v's Kind is not Interface or Pointer.
		argValue = reflect.New(m.ArgType).Elem()
	}
	return argValue
}

func (m *methodType) newReplyValue() reflect.Value {
	if m.ReplyType.Kind() != reflect.Pointer {
		log.Panic("xzrpc | reply must be a pointer type")
	}
	replyValue := reflect.New(m.ReplyType.Elem())

	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyValue.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))

	case reflect.Slice:
		replyValue.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}
	return replyValue
}

type service struct {
	// 结构体的名称，例如：service_test中的HelloService
	name string
	// 结构体的类型
	structType reflect.Type
	// 结构体的实例本身，后面需要作为参数0
	instance reflect.Value
	// 存储映射的结构体的所有方法
	method map[string]*methodType
}

// 创建service，入参是任意需要映射为服务的结构体实例
func newService(ins interface{}) *service {
	s := new(service)
	s.instance = reflect.ValueOf(ins)
	// ins可能是一个指针，通过Indirect方法返回指针指向的对象
	// 假设ins为指针，那么s.instance.Type()就是reflect.Ptr
	// (s.instance).Type().Name()，如果ins是对象，那么这个值就是""
	s.name = reflect.Indirect(s.instance).Type().Name()
	s.structType = reflect.TypeOf(ins)
	// 需要检查参数是否可导出，即符合func(t T) method(t1 T, t2 *T2) error的格式
	// 服务名的首字母是否大小
	if !ast.IsExported(s.name) {
		// 例如：HelloService，不是helloService
		log.Fatalf("xzrpc service | %s is invalid service name", s.name)
	}
	s.registerMethods()
	return s
}

func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)

	for i := 0; i < s.structType.NumMethod(); i++ {
		method := s.structType.Method(i)

		// prepare check，不满足条件的话，则该方法无法通过rpc调用
		// 入参0：结构体实例本身；入参1：参数t1；入参2：参数t2
		// 返回值0：error类型
		if method.Type.NumIn() != 3 || method.Type.NumOut() != 1 {
			continue
		}
		// 检查返回值，前面调用newArgValue和newReplyValue的时候，已经检查过参数
		if method.Type.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}

		argType, replyType := method.Type.In(1), method.Type.In(2)
		// 判断方法类型
		if !utils.IsExportedOrBuiltinType(argType) || !utils.IsExportedOrBuiltinType(replyType) {
			continue
		}

		s.method[method.Name] = &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("xzrpc server | register %s.%s\n", s.name, method.Name)

	}
}

// 通过反射值调用方法
func (s *service) call(m *methodType, argValue, replyValue reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)

	// 调用原方法，例如：调用helloService.Hello
	f := m.method.Func
	returnValues := f.Call([]reflect.Value{s.instance, argValue, replyValue})

	err := returnValues[0].Interface()
	if err != nil {
		return err.(error)
	}
	return nil
}
