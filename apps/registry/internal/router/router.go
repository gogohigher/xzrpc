package router

import "github.com/gin-gonic/gin"

func RegisterRouter(engine *gin.Engine) {
	apiGroup := engine.Group("_xzrpc_")
	registerApiRouter(apiGroup)
}

func registerApiRouter(group *gin.RouterGroup) {
	group.GET("ping", Ping)
	group.POST("registry", Register)
	group.POST("get", Get)
}
