package models

type RespMeta struct {
	ErrCode int     `json:"errCode"`
	ErrMsg  string  `json:"errMsg,omitempty"`
	Paging  *Paging `json:"paging,omitempty"`
}

type RespModel struct {
	Meta RespMeta    `json:"meta"`
	Data interface{} `json:"data,omitempty"`
}

type Paging struct {
	Page  int `json:"page"`  //当前页
	Size  int `json:"size"`  //每页数量
	Total int `json:"total"` //总数量
}

func NewPaging(page, size, total int) *Paging {
	return &Paging{
		Page:  page,
		Size:  size,
		Total: total,
	}
}

func NewMetaPaging(page, size, total int) RespMeta {
	return RespMeta{
		ErrCode: 0,
		ErrMsg:  "",
		Paging:  NewPaging(page, size, total),
	}
}

func NewPagingRespModel(data interface{}, page, size, total int) *RespModel {
	return &RespModel{
		Meta: NewMetaPaging(page, size, total),
		Data: data,
	}
}
