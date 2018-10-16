package routers

import (
	"net/http"

	"guoshi/routers/base"

	"github.com/gin-gonic/gin"
)

type PingRouter struct {
	base.BaseRouter
}

func NewPingRouter() *PingRouter {
	return &PingRouter{}
}

func (s *PingRouter) Load(routerGroup *gin.RouterGroup) {
	routerGroup.HEAD("", pingHandler)
	routerGroup.GET("", pingHandler)
	routerGroup.POST("", pingHandler)
	routerGroup.GET("/ping", pingHandler)
}

func pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
