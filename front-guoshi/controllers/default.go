package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type OrderResp struct {
	Meta struct {
		ErrCode int    `json:"errCode"`
		ErrMsg  string `json:"errMsg"`
	} `json:"meta"`
	Data []Order `json:"data"`
}

type UserResp struct {
	Meta struct {
		ErrCode int    `json:"errCode"`
		ErrMsg  string `json:"errMsg"`
	} `json:"meta"`
	Data User `json:"data"`
}

type ProjectResp struct {
	Meta struct {
		ErrCode int    `json:"errCode"`
		ErrMsg  string `json:"errMsg"`
	} `json:"meta"`
	Data []Project `json:"data"`
}
type JiShiResp struct {
	Meta struct {
		ErrCode int    `json:"errCode"`
		ErrMsg  string `json:"errMsg"`
	} `json:"meta"`
	Data struct {
		UserInfo     User          `json:"UserInfo"`
		UserProjects []UserProject `json:"UserProjects"`
	} `json:"data"`
}

type UserProject struct {
	Id          string  `json:"id"`
	UserId      string  `json:"userId"`
	ProjectId   string  `json:"projectId"`
	ProjectName string  `json:"projectName"`
	Level       int     `json:"level"`  // 1 高级 2 中级 3 初级
	ComeLv      float64 `json:"comeLv"` //提成点
}

type User struct {
	Id     string `json:"id"`
	Type   int    `json:"type"` // 1 技师 2 前台 3 管理员
	Number string `json:"number"`
	Name   string `json:"name"`
	Pwd    string `json:"pwd"`
	Status string `json:"status"`
}

type Order struct {
	Id         string    `json:"id"`
	UserId     string    `json:"userId"`
	UserNumber string    `json:"userNumber"`
	IsMember   int       `json:"isMember"`  // 0 非会员 1 是会员
	MustMoney  int       `json:"mustMoney"` //应收收款
	RealMoney  int       `json:"realMoney"` //实际收款
	Projects   []Project `json:"projects"`
	PayMethod  int       `json:"payMethod"`  //收款方式 1 现金 2 微信 3 支付宝
	Commission float64   `json:"commission"` //提成
	CreatedAt  int64     `json:"createdAt"`
	PayAt      int64     `json:"payAt"`
	Status     string    `json:"status"` // not-pay finish
}

type Project struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Number      string `json:"number"`
	Money       int    `json:"money"`
	IsDianZhong int    `json:"isDianZhong"` // 0 不是点钟 1 是点钟
	Status      string `json:"status"`
}

type JsShiOrder struct {
	Id         int     `json:"id"`
	UserNumber string  `json:"userNumber"`
	ProjectStr string  `json:"projectStr"`
	Commission float64 `json:"commission"`
	PayAt      string  `json:"payAt"`
}

type QianTaiOrder struct {
	Id          int    `json:"id"`
	OrderNumber string `json:"orderNumber"`
	OrderId     string `json:"orderId"`
	ProjectId   string `json:"projectId"`
	UserNumber  string `json:"userNumber"`
	ProjectName string `json:"projectName"`
	Money       int    `json:"money"`
	CreatedAt   string `json:"createdAt"`
}

type JiShiController struct {
	beego.Controller
}

type JiShiNotPayController struct {
	beego.Controller
}

type JiShiPayController struct {
	beego.Controller
}

type UserLoginController struct {
	beego.Controller
}

type OrderIndexController struct {
	beego.Controller
}

type OrderQianTaiController struct {
	beego.Controller
}

type OrderJishiController struct {
	beego.Controller
}

type OrderProjectController struct {
	beego.Controller
}

type OrderBaoBiaoController struct {
	beego.Controller
}

type BaoBiaoReq struct {
	Begin      int64
	End        int64
	UserNumber []string
}

func (c *OrderBaoBiaoController) Get() {
	c.TplName = "baobiao.html"
}

func (c *OrderJishiController) Get() {
	userNumber := c.Ctx.Request.URL.Query().Get("userNumber")
	body, err := HttpGet(fmt.Sprintf("http://47.94.140.226/guoshi/api/v1/user/user-pro?number=%s", userNumber))
	if err != nil {
		return
	}
	var data JiShiResp
	json.Unmarshal(body, &data)

	c.Data["UserInfo"] = data.Data.UserInfo
	c.Data["UserProjects"] = data.Data.UserProjects

	body, err = HttpGet("http://47.94.140.226/guoshi/api/v1/project/list")
	if err != nil {
		return
	}
	var dataPro ProjectResp
	json.Unmarshal(body, &dataPro)
	c.Data["Project"] = dataPro.Data
	c.TplName = "jishi.html"
}

func (c *OrderProjectController) Get() {
	body, err := HttpGet("http://47.94.140.226/guoshi/api/v1/project/list")
	if err != nil {
		return
	}
	var data ProjectResp
	json.Unmarshal(body, &data)
	c.Data["Project"] = data.Data
	c.TplName = "project.html"
}

func (c *UserLoginController) Get() {
	c.TplName = "login.html"
}

func (c *OrderQianTaiController) Get() {
	userNumber := c.Ctx.Request.URL.Query().Get("userNumber")
	fmt.Println(userNumber)
	body, err := HttpGet(fmt.Sprintf("http://47.94.140.226/guoshi/api/v1/order/user?number=%s", userNumber))
	if err != nil {
		return
	}
	var data OrderResp
	json.Unmarshal(body, &data)

	var qianTaiOrders []*QianTaiOrder
	for k, v := range data.Data {
		orderNumber := "10000-" + fmt.Sprintf("%d", k)
		for _, n := range v.Projects {
			qiantaiOrder := &QianTaiOrder{
				Id:          k + 1,
				OrderNumber: orderNumber,
				OrderId:     v.Id,
				ProjectId:   n.Id,
				UserNumber:  userNumber,
				ProjectName: n.Name,
				Money:       n.Money,
				CreatedAt:   time.Unix(v.CreatedAt, 0).Format("2006-01-02 15:04:05"), //设置时间戳 使用模板格式化为日期字符串
			}
			qianTaiOrders = append(qianTaiOrders, qiantaiOrder)
		}

	}
	c.Data["QianTaiOrder"] = qianTaiOrders
	c.TplName = "qiantai.html"
}

func (c *OrderIndexController) Get() {
	userNumber := c.Ctx.Request.URL.Query().Get("userNumber")
	pwd := c.Ctx.Request.URL.Query().Get("pwd")

	body, err := HttpGet(fmt.Sprintf("http://47.94.140.226/guoshi/api/v1/user/user/login?userNumber=%s&pwd=%s", userNumber, pwd))
	if err != nil {
		return
	}
	var data UserResp
	json.Unmarshal(body, &data)
	fmt.Println(data.Data.Type)
	c.Data["UserType"] = data.Data.Type
	c.TplName = "order.html"
}

func (c *JiShiController) Get() {
	c.TplName = "index.html"
}

func (c *JiShiNotPayController) Get() {
	userNumber := c.Ctx.Request.URL.Query().Get("userNumber")
	fmt.Println(userNumber)
	body, err := HttpGet(fmt.Sprintf("http://47.94.140.226/guoshi/api/v1/order/user?number=%s", userNumber))
	if err != nil {
		return
	}
	var data OrderResp
	json.Unmarshal(body, &data)

	var jsShiOrders []JsShiOrder
	for k, v := range data.Data {
		projectStr := ""
		for _, n := range v.Projects {
			projectStr += n.Name + "/"
		}

		var jsShiOrder JsShiOrder
		jsShiOrder.Id = k + 1
		jsShiOrder.UserNumber = userNumber
		jsShiOrder.ProjectStr = projectStr
		jsShiOrder.Commission = v.Commission
		dataTimeStr := time.Unix(v.CreatedAt, 0).Format("2006-01-02 15:04:05") //设置时间戳 使用模板格式化为日期字符串
		jsShiOrder.PayAt = dataTimeStr
		jsShiOrders = append(jsShiOrders, jsShiOrder)
	}
	fmt.Println(len(jsShiOrders))
	c.Data["JsShiOrder"] = jsShiOrders
	c.TplName = "notpay.html"
}

func (c *JiShiPayController) Get() {
	userNumber := c.Ctx.Request.URL.Query().Get("userNumber")
	day := c.Ctx.Request.URL.Query().Get("day")
	month := c.Ctx.Request.URL.Query().Get("month")

	var begin int64
	var end int64
	now := time.Now()
	if day != "" {
		begin = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).Unix()
		end = now.Unix()
	}
	if month != "" {
		begin = time.Date(now.Year(), now.Month(), 0, 0, 0, 0, 0, time.Local).Unix()
		end = now.Unix()
	}
	body, err := HttpGet(fmt.Sprintf("http://47.94.140.226/guoshi/api/v1/order/wage?begin=%s&end=%s&userNumber=%s", strconv.Itoa(int(begin)), strconv.Itoa(int(end)), userNumber))
	if err != nil {
		return
	}
	fmt.Println(string(body))
	var data OrderResp
	json.Unmarshal(body, &data)

	var jsShiOrders []JsShiOrder
	var dianzhong int
	var total float64
	for k, v := range data.Data {
		projectStr := ""
		for _, n := range v.Projects {
			if n.IsDianZhong == 1 {
				projectStr += n.Name + "-点钟/"
				dianzhong += 1
			} else {
				projectStr += n.Name + "/"
			}
		}

		var jsShiOrder JsShiOrder
		jsShiOrder.Id = k + 1
		jsShiOrder.UserNumber = userNumber
		jsShiOrder.ProjectStr = projectStr
		jsShiOrder.Commission = v.Commission
		dataTimeStr := time.Unix(v.PayAt, 0).Format("2006-01-02 15:04:05") //设置时间戳 使用模板格式化为日期字符串
		jsShiOrder.PayAt = dataTimeStr
		jsShiOrders = append(jsShiOrders, jsShiOrder)
		total += v.Commission
	}

	fmt.Println(len(jsShiOrders))
	c.Data["UserNumber"] = userNumber
	c.Data["JsShiOrder"] = jsShiOrders
	c.Data["DianZhong"] = dianzhong
	c.Data["Total"] = total
	c.TplName = "pay.html"
}

func HttpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("[http get] status err %s, %d\n", url, resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}
