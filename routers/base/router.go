package base

import "github.com/gin-gonic/gin"

type Router interface {
	Register(routeName string, child Router)
	Load(routeGroup *gin.RouterGroup)
}
