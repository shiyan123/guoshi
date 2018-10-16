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
)

type UserRouter struct {
	base.BaseRouter
}

func NewUserRouter() *UserRouter {
	return new(UserRouter)
}

func (u *UserRouter) Load(group *gin.RouterGroup) {
	group.GET("test", testHandler)

	group.POST("", CreatedUserHandler)
	group.POST("/update", updateUserHandler)
	group.GET("", getUserHandler)
	group.GET("/list", getUserListHandler)

	group.POST("/user-pro", createdUserHandler)
	group.POST("/user-pro/update", updateUserProHandler)
	group.GET("/user-pro", getUserProHandler)

	group.GET("/user/login", getUserLoginHandler)
}

func testHandler(c *gin.Context) {
	utils.JSON("success", c)
	return
}

func getUserHandler(c *gin.Context) {
	number := c.Query("number")
	if number == "" {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	var user models.User
	err := app.GetApp().Mongo.C(models.User_C).Find(bson.M{"number": number}).One(&user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorGetFail)
		return
	}
	utils.JSON(user, c)
	return
}

func getUserListHandler(c *gin.Context) {
	var users []models.User
	err := app.GetApp().Mongo.C(models.User_C).Find(nil).All(&users)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorGetFail)
		return
	}
	utils.JSON(users, c)
	return
}

func CreatedUserHandler(c *gin.Context) {
	var req models.User
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	req.Id = bson.NewObjectId()
	req.Status = "default"
	err := app.GetApp().Mongo.C(models.User_C).Insert(&req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorCreatedFail)
		return
	}
	utils.JSON("success", c)
	return
}

func updateUserHandler(c *gin.Context) {
	req := map[string]interface{}{}
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	if _, ok := req["id"]; !ok {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	info, err := app.GetApp().Mongo.C(models.User_C).Upsert(bson.M{"_id": bson.ObjectIdHex(req["id"].(string))}, req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorUpdateFail)
		return
	}
	fmt.Println(info.UpsertedId)
	utils.JSON("success", c)
	return
}

func getUserProHandler(c *gin.Context) {
	number := c.Query("number")
	if number == "" {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	var user models.User
	err := app.GetApp().Mongo.C(models.User_C).Find(bson.M{"number": number}).One(&user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorGetFail)
		return
	}
	var userProjects []models.UserProject
	err = app.GetApp().Mongo.C(models.User_Pro_C).Find(bson.M{"userId": user.Id.Hex()}).All(&userProjects)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorGetFail)
		return
	}
	var resp models.ResponseUserPro
	resp.UserInfo = user
	resp.UserProjects = userProjects

	utils.JSON(resp, c)
	return

}

func createdUserHandler(c *gin.Context) {
	var req models.RequestUserPro
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	var user models.User
	err := app.GetApp().Mongo.C(models.User_C).Find(bson.M{"number": req.UserNumber}).One(&user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorGetFail)
		return
	}
	for _, v := range req.UserProjects {
		v.Id = bson.NewObjectId()
		v.UserId = user.Id.Hex()
		err := app.GetApp().Mongo.C(models.User_Pro_C).Insert(&v)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorCreatedFail)
			return
		}
	}
	utils.JSON("success", c)
	return
}

func updateUserProHandler(c *gin.Context) {
	var req models.RequestUserPro
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	var user models.User
	err := app.GetApp().Mongo.C(models.User_C).Find(bson.M{"number": req.UserNumber}).One(&user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorGetFail)
		return
	}
	for _, v := range req.UserProjects {
		info, err := app.GetApp().Mongo.C(models.User_Pro_C).Upsert(bson.M{"_id": v.Id}, v)
		if err != nil {
			fmt.Println(err)
			c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorUpdateFail)
			return
		}
		fmt.Println(info.UpsertedId)
	}
	utils.JSON("success", c)
	return
}

func getUserLoginHandler(c *gin.Context) {
	userNumber := c.Query("userNumber")
	pwd := c.Query("pwd")
	var user *models.User
	err := app.GetApp().Mongo.C(models.User_C).Find(bson.M{"number": userNumber, "pwd": pwd}).One(&user)
	if err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorGetFail)
		return
	}
	if user == nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorGetFail)
		return
	}
	utils.JSON(user, c)
	return
}
