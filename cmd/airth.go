package cmd

import (
	"time"
)

type Arith int

type Args struct {
	A, B int
}

func (a *Arith) Multiply(args Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

func (a *Arith) MultiplySleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.A))
	*reply = args.A * args.B
	return nil
}
