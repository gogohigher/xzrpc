package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gogohigher/xzrpc/apps/registry/internal/model"
	"github.com/gogohigher/xzrpc/apps/registry/internal/registry"
	"log"
	"net/http"
)

func Ping(ctx *gin.Context) {
	fmt.Println("ping...ping...ping...")
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

func Register(ctx *gin.Context) {
	fmt.Println("registry | start Register.")

	var req model.RegisterRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}
	// req -> serverItem
	item := registry.NewServerItem(&req)

	// 注册
	_, err := registry.GlobalRegistry.Register(item)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
	log.Println("registry | Register success")

}
