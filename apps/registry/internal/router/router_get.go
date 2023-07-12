package router

import (
	"github.com/gin-gonic/gin"
	"github.com/gogohigher/xzrpc/apps/registry/internal/model"
	"github.com/gogohigher/xzrpc/apps/registry/internal/registry"
	"log"
	"net/http"
	"strings"
	"time"
)

func Get(ctx *gin.Context) {
	var req model.GetRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}
	servers, err := registry.GlobalRegistry.GetServer(req.AppId, req.Env)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    -1,
			"message": "获取地址失败",
		})
		return
	}

	aliveServers := make([]string, 0)

	for _, server := range servers {
		if registry.GlobalRegistry.Timeout == 0 || time.Duration(time.Now().Unix()-server.RegTimestamp) <= registry.GlobalRegistry.Timeout {
			aliveServers = append(aliveServers, server.Address)
		} else {
			// TODO 这里不要删除，感觉不应该在这里处理
			// 1. 删除app中的items
			// 2. 如果app中的items为空，删除registry中的apps
			//delete(r.servers, addr)
		}
	}
	ctx.Header("X-Xzrpc-Servers", strings.Join(aliveServers, ","))

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    servers,
	})

	log.Println("registry | GetServer success")
}
