package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
	"guoshi/app"
	"guoshi/consts"
	"guoshi/models"
	"guoshi/routers/base"
	"guoshi/utils"
	"net/http"
	"strconv"
	"time"
)

type OrderRouter struct {
	base.BaseRouter
}

func NewOrderRouter() *OrderRouter {
	return new(OrderRouter)
}

func (u *OrderRouter) Load(group *gin.RouterGroup) {
	group.POST("", CreatedOrderHandler)
	group.POST("/pay", payOrderHandler)
	group.GET("/user", getUserOrderHandler)
	group.GET("/wage", getUserWageHandler)

	group.POST("/revenue", getRevenueHandler)
	group.GET("/deleted", deletedOrderHandler)

	group.GET("/download", DownloadHandler)

	group.GET("/order-list", OrderListHandler)
	group.GET("/order-id", OrderDetailHandler)
	group.POST("/order-update", OrderUpdateHandler)
}

func CreatedOrderHandler(c *gin.Context) {
	var req models.OrderReq
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	var user models.User
	err := app.GetApp().Mongo.C(models.User_C).Find(bson.M{"number": req.UserNumber}).One(&user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorCreatedFail)
		return
	}
	var pros []models.Project
	var mustMoney int
	var commission float64
	for _, v := range req.ProjectIds {
		var pro models.Project
		err := app.GetApp().Mongo.C(models.Project_C).Find(bson.M{"_id": bson.ObjectIdHex(v)}).One(&pro)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorCreatedFail)
			return
		}
		var userPro models.UserProject
		err = app.GetApp().Mongo.C(models.User_Pro_C).Find(bson.M{"userId": user.Id.Hex(), "projectId": v}).One(&userPro)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorCreatedFail)
			return
		}

		commission = commission + userPro.ComeLv
		mustMoney += pro.Money
		pros = append(pros, pro)
	}

	order := &models.Order{
		Id:         bson.NewObjectId(),
		UserId:     user.Id.Hex(),
		UserNumber: user.Number,
		MustMoney:  mustMoney,
		Projects:   pros,
		Commission: commission,
		CreatedAt:  time.Now().Unix(),
		Status:     "not-pay",
	}
	err = app.GetApp().Mongo.C(models.Order_C).Insert(order)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorCreatedFail)
		return
	}
	utils.JSON("success", c)
	return
}

func payOrderHandler(c *gin.Context) {
	var req models.PayOrder
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}

	for _, v := range req.DianZhongProJectIds {
		err := app.GetApp().Mongo.C(models.Order_C).Update(bson.M{"_id": bson.ObjectIdHex(req.OrderId), "projects._id": bson.ObjectIdHex(v)}, bson.M{"$set": bson.M{"projects.$.isDianZhong": 1}})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorUpdateFail)
			return
		}
	}
	realMoney := req.HuiYuan + req.WeiXin + req.ZhiFuBao + req.XianJin +req.TuanGou
	err := app.GetApp().Mongo.C(models.Order_C).Update(bson.M{"_id": bson.ObjectIdHex(req.OrderId)},
		bson.M{"$set": bson.M{"isMember": req.IsMember, "huiYuan": req.HuiYuan, "weiXin": req.WeiXin, "zhiFuBao": req.ZhiFuBao, "xianJin": req.XianJin, "tuanGou": req.TuanGou, "realMoney": realMoney, "payMethod": req.PayMethod,
			"payAt": time.Now().Unix(), "status": "finish"}})

	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorUpdateFail)
		return
	}
	utils.JSON("success", c)
	return
}

func getUserOrderHandler(c *gin.Context) {
	userNumber := c.Query("number")
	if userNumber == "" {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	var orders []models.Order
	err := app.GetApp().Mongo.C(models.Order_C).Find(bson.M{"userNumber": userNumber, "status": "not-pay"}).All(&orders)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorGetFail)
		return
	}
	utils.JSON(orders, c)
	return
}

func getUserWageHandler(c *gin.Context) {
	beginStr := c.Query("begin")
	endStr := c.Query("end")
	userNumber := c.Query("userNumber")

	if beginStr == "" || endStr == "" || userNumber == "" {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	begin, err := strconv.Atoi(beginStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	end, err := strconv.Atoi(endStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	var orders []models.Order
	err = app.GetApp().Mongo.C(models.Order_C).Find(bson.M{"userNumber": userNumber, "status": "finish", "payAt": bson.M{"$gte": int64(begin), "$lte": int64(end)}}).All(&orders)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorGetFail)
		return
	}
	var mustMoney int
	var realMoney float64
	var dianZhong int
	var commMoney float64
	for _, v := range orders {
		mustMoney += v.MustMoney
		realMoney += v.RealMoney
		for _, n := range v.Projects {
			if n.IsDianZhong == 1 {
				dianZhong += 1
			}
		}
		commMoney += v.Commission
	}
	utils.JSONPAY(orders, mustMoney, realMoney, dianZhong, commMoney, c)
	return
}

func getRevenueHandler(c *gin.Context) {
	var req models.Revenue
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	var orders []models.Order
	if len(req.UserNumber) == 0 {
		err := app.GetApp().Mongo.C(models.Order_C).Find(bson.M{"status": "finish", "payAt": bson.M{"$gte": req.Begin, "$lte": req.End}}).All(&orders)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorGetFail)
			return
		}
	} else {
		for _, v := range req.UserNumber {
			var userOrders []models.Order
			err := app.GetApp().Mongo.C(models.Order_C).Find(bson.M{"userNumber": v, "status": "finish", "payAt": bson.M{"$gte": req.Begin, "$lte": req.End}}).All(&userOrders)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorGetFail)
				return
			}
			orders = append(orders, userOrders...)
		}
	}
	var mustMoney int
	var realMoney float64
	var dianZhong int
	var commMoney float64
	for _, v := range orders {
		mustMoney += v.MustMoney
		realMoney += v.RealMoney
		for _, n := range v.Projects {
			if n.IsDianZhong == 1 {
				dianZhong += 1
			}
		}
		commMoney += v.Commission
	}
	utils.JSONPAY(orders, mustMoney, realMoney, dianZhong, commMoney, c)
	return
}

func deletedOrderHandler(c *gin.Context) {
	orderId := c.Query("orderId")
	projectId := c.Query("projectId")
	if orderId == "" || projectId == "" {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	err := app.GetApp().Mongo.C(models.Order_C).Update(bson.M{"_id": bson.ObjectIdHex(orderId)}, bson.M{"$pop": bson.M{"projects": bson.M{"id": bson.ObjectIdHex(projectId)}}})
	fmt.Println(err)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorUpdateFail)
		return
	}
	utils.JSON("success", c)
	return
}

func DownloadHandler(c *gin.Context) {
	BeginStr := c.Query("begin")
	EndStr := c.Query("end")
	beginTime, err := time.ParseInLocation("2006-01-02", BeginStr, time.Local)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	endTime, err := time.ParseInLocation("2006-01-02", EndStr, time.Local)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	resp, err := download(beginTime.Unix(), endTime.Unix())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}

	body, err := utils.GeneralXlsxGenerator(resp)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}

	c.Header("X-Powered-By", "Hortor Server")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet;charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment;filename=%s_%s_%s.xlsx", "data", "guoshi", "stats"))
	c.Header("Content-Length", fmt.Sprintf("%d", body.Len()))
	c.Writer.Write(body.Bytes())
	return

}

func download(begin, end int64) (newResp []*models.StatsData, err error) {
	beginTime := time.Unix(int64(begin), 0)
	endTime := time.Unix(int64(end), 0)
	var users []models.User

	fmt.Println(">>>>>>>>>>查询用户开始时间", time.Now().Unix())
	err = app.GetApp().Mongo.C(models.User_C).Find(nil).All(&users)
	if err != nil {
		return
	}
	fmt.Println(">>>>>>>>>>查询用户结束时间", time.Now().Unix())

	var userIds []string
	userIdMap := make(map[string]string, 0)
	for _, v := range users {
		if v.Type == 3 {
			userIdMap[v.Id.Hex()] = v.Number
			userIds = append(userIds, v.Id.Hex())
		}
	}

	for ts := beginTime; ts.Before(endTime); ts = ts.Add(24 * time.Hour) {
		begin := ts
		end := ts.Add(24 * time.Hour)

		var orders []models.Order
		err := app.GetApp().Mongo.C(models.Order_C).Find(bson.M{"userId": bson.M{"$in": userIds}, "status": "finish", "payAt": bson.M{"$gte": begin.Unix(), "$lte": end.Unix()}}).Sort("userNumber", "payAt").All(&orders)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		for _, n := range orders {
			for l, p := range n.Projects {
				if l == 0 {
					stats := &models.StatsData{
						TimeStr:     ts.Format("2006-01-02"),
						UserNumber:  userIdMap[n.UserId],
						ProJectName: p.Name,
						PayAt:       time.Unix(n.PayAt, 0).Format("2006-01-02 15:04:05"),
						Commission:  n.Commission,
						HuiYuan:     n.HuiYuan,
						WeiXin:      n.WeiXin,
						ZhiFuBao:    n.ZhiFuBao,
						XianJin:     n.XianJin,
						TuanGou:     n.TuanGou,
						RealMonry:   n.RealMoney,
					}
					newResp = append(newResp, stats)
				} else {
					stats := &models.StatsData{
						TimeStr:     ts.Format("2006-01-02"),
						UserNumber:  userIdMap[n.UserId],
						ProJectName: p.Name,
						PayAt:       time.Unix(n.PayAt, 0).Format("2006-01-02 15:04:05"),
						Commission:  0,
						HuiYuan:     0,
						WeiXin:      0,
						ZhiFuBao:    0,
						XianJin:     0,
						TuanGou:     0,
						RealMonry:   0,
					}
					newResp = append(newResp, stats)
				}

			}
		}
		//for _, v := range users {
		//	if v.Type != 3 {
		//		continue
		//	}
		//	var orders []models.Order
		//	fmt.Println(">>>>>>>>>>查询订单开始时间", time.Now().Unix())
		//	err := app.GetApp().Mongo.C(models.Order_C).Find(bson.M{"userId": v.Id.Hex(), "status": "finish", "payAt": bson.M{"$gte": begin.Unix(), "$lte": end.Unix()}}).All(&orders)
		//	if err != nil {
		//		fmt.Println(err)
		//		return nil, err
		//	}
		//	fmt.Println(">>>>>>>>>>查询订单结束时间", time.Now().Unix())
		//	for _, n := range orders {
		//		for l, p := range n.Projects {
		//			if l == 0 {
		//				stats := &models.StatsData{
		//					TimeStr:     ts.Format("2006-01-02"),
		//					UserNumber:  v.Number,
		//					ProJectName: p.Name,
		//					PayAt:       time.Unix(n.PayAt, 0).Format("2006-01-02 15:04:05"),
		//					Commission:  n.Commission,
		//					HuiYuan:     n.HuiYuan,
		//					WeiXin:      n.WeiXin,
		//					ZhiFuBao:    n.ZhiFuBao,
		//					XianJin:     n.XianJin,
		//					TuanGou:     n.TuanGou,
		//					RealMonry:   n.RealMoney,
		//				}
		//				newResp = append(newResp, stats)
		//			} else {
		//				stats := &models.StatsData{
		//					TimeStr:     ts.Format("2006-01-02"),
		//					UserNumber:  v.Number,
		//					ProJectName: p.Name,
		//					PayAt:       time.Unix(n.PayAt, 0).Format("2006-01-02 15:04:05"),
		//					Commission:  0,
		//					HuiYuan:     0,
		//					WeiXin:      0,
		//					ZhiFuBao:    0,
		//					XianJin:     0,
		//					TuanGou:     0,
		//					RealMonry:   0,
		//				}
		//				newResp = append(newResp, stats)
		//			}
		//
		//		}
		//	}
		//}
	}
	var huiyuan float64
	var weixin float64
	var zhifubao float64
	var xianjin float64
	var tuangou float64
	var commission float64
	var realMoney float64
	for _, v := range newResp {
		huiyuan += v.HuiYuan
		weixin += v.WeiXin
		zhifubao += v.ZhiFuBao
		xianjin += v.XianJin
		tuangou += v.TuanGou
		commission += v.Commission
		realMoney += v.RealMonry
	}
	stats := &models.StatsData{
		TimeStr:     "",
		UserNumber:  "",
		ProJectName: "",
		PayAt:       "",
		Commission:  commission,
		HuiYuan:     huiyuan,
		WeiXin:      weixin,
		ZhiFuBao:    zhifubao,
		XianJin:     xianjin,
		TuanGou:     tuangou,
		RealMonry:   realMoney,
	}
	newResp = append(newResp, stats)
	return
}

func OrderListHandler(c *gin.Context){
	timeStr := c.Query("timeStr")
	userNumber := c.Query("userNumber")
	if timeStr == "" || userNumber == "" {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	beginTime, err := time.Parse("2006-01-02", timeStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	endTime := beginTime.AddDate(0, 0, 1)
	var orders []models.Order
	err = app.GetApp().Mongo.C(models.Order_C).
		Find(bson.M{"userNumber":userNumber, "createdAt":bson.M{"$gte":beginTime.Unix(),"$lte":endTime.Unix()}}).All(&orders)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorGetFail)
		return
	}
	utils.JSON(orders, c)
	return
}

func OrderDetailHandler(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	var order models.Order
	err := app.GetApp().Mongo.C(models.Order_C).
		Find(bson.M{"_id":bson.ObjectIdHex(id)}).One(&order)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorGetFail)
		return
	}
	order.CreatedAtStr = time.Unix(order.CreatedAt, 0).Format("2006-01-02 15:04:05")
	order.PayAtStr = "暂未支付"
	if order.PayAt != 0 {
		order.PayAtStr = time.Unix(order.PayAt, 0).Format("2006-01-02 15:04:05")
	}
	utils.JSON(order, c)
	return
}

func OrderUpdateHandler(c *gin.Context) {
	var projects models.OrderUpdate
	if err := c.BindJSON(&projects); err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	var mustMoney int
	var commission float64
	var pros []models.Project
	for _, v := range projects.Projects {
		var pro models.Project
		err := app.GetApp().Mongo.C(models.Project_C).Find(bson.M{"_id": bson.ObjectIdHex(v.ProjectId)}).One(&pro)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorCreatedFail)
			return
		}
		var userPro models.UserProject
		err = app.GetApp().Mongo.C(models.User_Pro_C).Find(bson.M{"userId": projects.UserId, "projectId": v.ProjectId}).One(&userPro)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorCreatedFail)
			return
		}
		if v.IsDianZhong == true {
			pro.IsDianZhong = 1
		}
		pro.Status = "default"
		pros =append(pros, pro)
		commission = commission + userPro.ComeLv
		mustMoney += pro.Money
	}
	err := app.GetApp().Mongo.C(models.Order_C).
		Update(bson.M{"_id":bson.ObjectIdHex(projects.OrderId)},
		bson.M{"$set":bson.M{"mustMoney":mustMoney,"commission":commission,"projects":pros}})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorUpdateFail)
		return
	}
	utils.JSON("success", c)
	return
}