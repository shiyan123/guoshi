package utils

import (
	"github.com/gin-gonic/gin"
	model "guoshi/models"
	"net/http"
)

func RespWriteErrorWithCode(errCode int, msg string) *model.RespModel {
	return &model.RespModel{
		Meta: model.RespMeta{ErrCode: errCode, ErrMsg: msg},
	}
}

func RespWriteError(resp *model.RespModel, msg string) *model.RespModel {
	return &model.RespModel{
		Meta: model.RespMeta{ErrCode: resp.Meta.ErrCode, ErrMsg: msg},
	}
}

func RespWriteData(resp *model.RespModel, data interface{}) *model.RespModel {
	return &model.RespModel{
		Meta: resp.Meta,
		Data: data,
	}
}

func NewPaging(currentPage, pageSize, total int) *Paging {
	totalPage := total / pageSize
	if total%pageSize != 0 {
		totalPage += 1
	}

	return &Paging{
		Total:       total,
		TotalPage:   totalPage,
		CurrentPage: currentPage,
		PageSize:    pageSize,
	}
}
func JSON(data interface{}, c *gin.Context) {
	res := newResApi(data)
	c.JSON(http.StatusOK, res)
}

func JSONPAY(data interface{}, mustMoney int, realMoney float64, dianZhongCnt int, commMoney float64, c *gin.Context) {
	res := newResOrderApi(data, mustMoney, realMoney, dianZhongCnt, commMoney)
	c.JSON(http.StatusOK, res)
}

func JSONPaging(data interface{}, currentPage, pageSize, total int, c *gin.Context) {
	res := newPagingApi(data, currentPage, pageSize, total)
	c.JSON(http.StatusOK, res)
	return
}

type ResApi2 struct {
	Meta *Meta       `json:"meta"`
	Data interface{} `json:"data"`
}

type ResApi3 struct {
	Meta         *Meta       `json:"meta"`
	Data         interface{} `json:"data"`
	MustMoney    int         `json:"mustMoney"`
	RealMoney    float64     `json:"realMoney"`
	CommMoney    float64     `json:"commMoney"`
	DianZhongCnt int         `json:"dianZhongCnt"`
}

type Meta struct {
	ErrCode int     `json:"errCode"`
	ErrMsg  string  `json:"errMsg,omitempty"`
	Paging  *Paging `json:"paging"`
}

type Paging struct {
	Total       int `json:"total"`
	TotalPage   int `json:"totalPage"`
	CurrentPage int `json:"currentPage"`
	PageSize    int `json:"pageSize"`
}

func NewDefaultMetaApi() *Meta {
	return &Meta{
		ErrCode: 0,
		ErrMsg:  "success",
	}
}
func newResApi(data interface{}) *ResApi2 {
	return &ResApi2{
		Meta: NewDefaultMetaApi(),
		Data: data,
	}
}

func newResOrderApi(data interface{}, mustMoney int, realMoney float64, dianZhongCnt int, commMoney float64) *ResApi3 {
	return &ResApi3{
		Meta:         NewDefaultMetaApi(),
		Data:         data,
		MustMoney:    mustMoney,
		RealMoney:    realMoney,
		CommMoney:    commMoney,
		DianZhongCnt: dianZhongCnt,
	}
}

func NewMetaPagingApi(currentPage, pageSize, total int) *Meta {
	return &Meta{
		ErrCode: 0,
		ErrMsg:  "success",
		Paging:  NewPaging(currentPage, pageSize, total),
	}
}

func newPagingApi(data interface{}, currentPage, pageSize, total int) *ResApi2 {
	return &ResApi2{
		Meta: NewMetaPagingApi(currentPage, pageSize, total),
		Data: data,
	}
}
