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

type ProjectRouter struct {
	base.BaseRouter
}

func NewProjectRouter() *ProjectRouter {
	return new(ProjectRouter)
}

func (u *ProjectRouter) Load(group *gin.RouterGroup) {

	group.POST("", CreatedProjectHandler)
	group.POST("/update", updateProjectHandler)
	group.GET("", getProjectHandler)
	group.GET("/list", getProjectListHandler)

	group.GET("/deleted", deletedProjectHandler)

}

func getProjectHandler(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	var pro models.Project
	err := app.GetApp().Mongo.C(models.Project_C).Find(bson.M{"name": name}).One(&pro)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorGetFail)
		return
	}
	utils.JSON(pro, c)
	return
}

func getProjectListHandler(c *gin.Context) {
	var pros []models.Project
	err := app.GetApp().Mongo.C(models.Project_C).Find(nil).All(&pros)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorGetFail)
		return
	}
	utils.JSON(pros, c)
	return
}

func CreatedProjectHandler(c *gin.Context) {
	var req models.Project
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	req.Id = bson.NewObjectId()
	req.Status = "default"
	err := app.GetApp().Mongo.C(models.Project_C).Insert(&req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorCreatedFail)
		return
	}
	utils.JSON("success", c)
	return
}

func updateProjectHandler(c *gin.Context) {
	req := map[string]interface{}{}
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	if _, ok := req["id"]; !ok {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	info, err := app.GetApp().Mongo.C(models.Project_C).Upsert(bson.M{"_id": bson.ObjectIdHex(req["id"].(string))}, req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorUpdateFail)
		return
	}
	fmt.Println(info.UpsertedId)
	utils.JSON("success", c)
	return
}

func deletedProjectHandler(c *gin.Context) {
	projectId := c.Query("projectId")
	if projectId == "" {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorParamsInvalid)
		return
	}
	err := app.GetApp().Mongo.C(models.Project_C).Remove(bson.M{"_id": bson.ObjectIdHex(projectId)})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, consts.RespErrorUpdateFail)
		return
	}
	utils.JSON("success", c)
	return
}
