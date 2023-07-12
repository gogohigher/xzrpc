package model

// RegisterRequest 注册请求对象
type RegisterRequest struct {
	Env      string `json:"env" form:"env"`
	AppId    string `json:"appid" form:"appid"`
	Hostname string `json:"hostname" form:"hostname"`
	Address  string `json:"address" form:"address"` // TODO 后续改成[]string
}

// GetRequest 获取服务
type GetRequest struct {
	AppId string `json:"appid" form:"appid"`
	Env   string `json:"env" form:"env"`
}
