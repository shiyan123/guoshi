package v1

import "guoshi/routers/base"

type V1Router struct {
	base.BaseRouter
}

func NewV1Router() *V1Router {
	return (&V1Router{}).init()
}

func (r *V1Router) init() *V1Router {
	r.Register("/user", NewUserRouter())
	r.Register("/project", NewProjectRouter())
	r.Register("/order", NewOrderRouter())
	return r
}
