package routers

import (
	"github.com/astaxie/beego"
	"guoshi/front-guoshi/controllers"
)

func init() {
	beego.Router("/jishi", &controllers.JiShiController{})
	beego.Router("/jishi/notPay", &controllers.JiShiNotPayController{})
	beego.Router("/jishi/pay", &controllers.JiShiPayController{})

	beego.Router("/user/login", &controllers.UserLoginController{})
	beego.Router("/order/index", &controllers.OrderIndexController{})

	beego.Router("/order/qiantai", &controllers.OrderQianTaiController{})
	beego.Router("/order/jishi", &controllers.OrderJishiController{})
	beego.Router("/order/project", &controllers.OrderProjectController{})
	beego.Router("/order/baobiao", &controllers.OrderBaoBiaoController{})
}
