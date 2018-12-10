package main

import (
	"github.com/astaxie/beego"
	_ "guoshi/front-guoshi/routers"
)

func main() {
	beego.Run("0.0.0.0:8080")
}
