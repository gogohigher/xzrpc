package xzgin

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type GinServer struct {
	Engine *gin.Engine
}

func NewGinServer() *GinServer {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	return &GinServer{
		Engine: engine,
	}
}

func (gs *GinServer) Run(port int) {
	addr := ":" + strconv.Itoa(port)
	err := gs.Engine.Run(addr)
	if err != nil {
		panic(err)
	}
}

func (gs *GinServer) Use(m ...gin.HandlerFunc) gin.IRoutes {
	return gs.Engine.Use(m...)
}
