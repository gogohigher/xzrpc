package server

import (
	"github.com/gogohigher/xzrpc/apps/registry/internal/router"
	"github.com/gogohigher/xzrpc/pkg/xzgin"
)

type Server struct {
	ginServer *xzgin.GinServer
	//conf      *config.Config
}

func NewServer() *Server {
	s := &Server{
		ginServer: xzgin.NewGinServer(),
	}
	return s
}

func (s *Server) Run() {
	router.RegisterRouter(s.ginServer.Engine)
	s.ginServer.Run(9999)
}
