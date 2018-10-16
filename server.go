package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"guoshi/routers"
)

func Listen(address string, port int) error {
	r := gin.Default()

	routers.Load(r)

	return r.Run(fmt.Sprintf("%s:%d", address, port))
}
