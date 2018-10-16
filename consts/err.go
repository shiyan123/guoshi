package consts

import (
	"guoshi/utils"
)

var (
	ErrParamsInvalid       = utils.RespWriteErrorWithCode(10000, "ok")
	ErrUserNotFound        = utils.RespWriteErrorWithCode(10001, "用户或密码不对")
	RespErrorParamsInvalid = utils.RespWriteErrorWithCode(-100001, "参数错误")
	RespErrorCreatedFail   = utils.RespWriteErrorWithCode(-100002, "创建失败")
	RespErrorGetFail       = utils.RespWriteErrorWithCode(-100003, "查询失败")
	RespErrorUpdateFail    = utils.RespWriteErrorWithCode(-100004, "更新失败")
	RespErrorUpload        = utils.RespWriteErrorWithCode(-100005, "上传失败")
)
