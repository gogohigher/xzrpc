package message

import (
	"time"
)

type Arith int

func (a *Arith) Multiply(args ArithRequest, reply *ArithResponse) error {
	reply.C = args.A * args.B
	return nil
}

func (a *Arith) MultiplySleep(args ArithRequest, reply *ArithResponse) error {
	time.Sleep(time.Second * time.Duration(args.A))
	reply.C = args.A * args.B
	return nil
}
