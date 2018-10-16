package base

import "github.com/gin-gonic/gin"

type BaseRouter struct {
	RouteName    string
	childRouters map[string]Router
}

func (r *BaseRouter) Register(routeName string, child Router) {
	if r.childRouters == nil {
		r.childRouters = map[string]Router{}
	}

	r.childRouters[routeName] = child
}

func (r *BaseRouter) Load(routeGroup *gin.RouterGroup) {
	for routeName, router := range r.childRouters {
		group := routeGroup.Group(routeName)
		router.Load(group)
	}
}
