package models

import (
	"gopkg.in/mgo.v2/bson"
)

const (
	User_C     = "users"
	Project_C  = "project"
	Order_C    = "order"
	User_Pro_C = "user-pro"
)

type User struct {
	Id     bson.ObjectId `bson:"_id" json:"id"`
	Type   int           `bson:"type" json:"type"` // 1 管理员 2 前台 3 技师
	Number string        `bson:"number" json:"number"`
	Name   string        `bson:"name" json:"name"`
	Pwd    string        `bson:"pwd" json:"pwd"`
	Status string        `bson:"status" json:"status"`
}

type UserProject struct {
	Id          bson.ObjectId `bson:"_id" json:"id"`
	UserId      string        `bson:"userId" json:"userId"`
	ProjectId   string        `bson:"projectId" json:"projectId"`
	ProjectName string        `bson:"projectName" json:"projectName"`
	Level       int           `bson:"level" json:"level"`   // 1 高级 2 中级 3 初级
	ComeLv      float64       `bson:"comeLv" json:"comeLv"` //提成
}

type RequestUserPro struct {
	UserNumber   string        `json:"userNumber"`
	UserProjects []UserProject `json:"userProjects"`
}

type ResponseUserPro struct {
	UserInfo     User
	UserProjects []UserProject
}

type Project struct {
	Id          bson.ObjectId `bson:"_id" json:"id"`
	Name        string        `bson:"name" json:"name"`
	Number      string        `bson:"number" json:"number"`
	Money       int           `bson:"money" json:"money"`
	IsDianZhong int           `bson:"isDianZhong" json:"isDianZhong"` // 0 不是点钟 1 是点钟
	Status      string        `bson:"status" json:"status"`
}

type Order struct {
	Id         bson.ObjectId `bson:"_id" json:"id"`
	UserId     string        `bson:"userId" json:"userId"`
	UserNumber string        `bson:"userNumber" json:"userNumber"`
	IsMember   int           `bson:"isMember" json:"isMember"`   // 0 非会员 1 是会员
	MustMoney  int           `bson:"mustMoney" json:"mustMoney"` //应收收款
	RealMoney  float64       `bson:"realMoney" json:"realMoney"` //实际收款
	Projects   []Project     `bson:"projects" json:"projects"`
	Commission float64       `bson:"commission" json:"commission"` //提成
	CreatedAt  int64         `bson:"createdAt" json:"createdAt"`
	HuiYuan    float64       `bson:"huiYuan" json:"huiYuan"`
	WeiXin     float64       `bson:"weiXin" json:"weiXin"`
	ZhiFuBao   float64       `bson:"zhiFuBao" json:"zhiFuBao"`
	XianJin    float64       `bson:"xianJin" json:"xianJin"`
	TuanGou    float64       `bson:"tuanGou" json:"tuanGou"`
	PayAt      int64         `bson:"payAt" json:"payAt"`
	Status     string        `bson:"status" json:"status"` // not-pay finish
}

type OrderReq struct {
	UserNumber string   `json:"userNumber"`
	ProjectIds []string `json:"projectIds"`
}

type PayOrder struct {
	OrderId             string   `json:"orderId"`
	IsMember            int      `json:"isMember"`
	HuiYuan             float64  `json:"huiYuan"`
	WeiXin              float64  `json:"weiXin"`
	ZhiFuBao            float64  `json:"zhiFuBao"`
	XianJin             float64  `json:"xianJin"`
	TuanGou             float64  `json:"tuanGou"`
	RealMoney           float64  `json:"realMoney"`
	PayMethod           int      `json:"payMethod"`
	DianZhongProJectIds []string `json:"dianZhongProJectIds"`
}

type Revenue struct {
	Begin      int64    `json:"begin"`
	End        int64    `json:"end"`
	UserNumber []string `json:"userNumber"`
}

type StatsData struct {
	TimeStr     string  `json:"timeStr" xson:"日期"`
	UserNumber  string  `json:"userNumber" xson:"技师编号"`
	ProJectName string  `json:"proJectName" xson:"项目"`
	PayAt       string  `json:"payAt" xson:"结账时间"`
	Commission  float64 `json:"payAt" xson:"提成"`
	HuiYuan     float64 `json:"huiYuan" xson:"会员"`
	WeiXin      float64 `json:"weiXin" xson:"微信"`
	ZhiFuBao    float64 `json:"zhiFuBao" xson:"支付宝"`
	XianJin     float64 `json:"xianJin" xson:"现金"`
	TuanGou     float64 `json:"tuanGou" xson:"团购"`
	RealMonry   float64 `json:"realMonry" xson:"合计"`
}
