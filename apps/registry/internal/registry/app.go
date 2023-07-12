package registry

import (
	"github.com/gogohigher/xzrpc/apps/registry/internal/model"
	"sync"
	"time"
)

type App struct {
	appid string                 // 应用服务唯一标识
	items map[string]*ServerItem // <host, 服务实例>
	//latestTimestamp int64                  // 上一次更新时间
	lock sync.RWMutex
}

type ServerItem struct {
	Address string // 先写单个
	//Addresses []string // 服务实例的地址
	//start        time.Time // 开始时间，等同于注册时间，后续废弃掉
	Env          string // 服务环境，例如：online、dev、test
	AppId        string // 应用服务的唯一标识
	Hostname     string // 服务实例的唯一标识
	RegTimestamp int64  // 服务注册时间戳，等于start，单位秒
}

func NewApp(appid string) *App {
	app := &App{
		appid: appid,
		items: make(map[string]*ServerItem),
	}
	return app
}

func NewServerItem(req *model.RegisterRequest) *ServerItem {
	now := time.Now().Unix()
	ins := &ServerItem{
		AppId:        req.AppId,
		Env:          req.Env,
		Hostname:     req.Hostname,
		Address:      req.Address,
		RegTimestamp: now,
	}
	return ins
}

func (app *App) addServer(item *ServerItem) {
	app.lock.Lock()
	defer app.lock.Unlock()
	app.items[item.Hostname] = item
	// 应用服务有好几个服务实例，不能将某一个服务实例的最新更新时间当做app的最新更新时间
	//app.lastTimestamp = latestTimestamp
}

// deep copy
func (app *App) getServers() ([]*ServerItem, error) {
	app.lock.Lock()
	defer app.lock.Unlock()
	res := make([]*ServerItem, 0)
	for _, item := range app.items { // 此时可以不看域名，全部返回
		copyItem := copyServerItem(item)
		res = append(res, copyItem)
	}
	return res, nil
}

func copyServerItem(in *ServerItem) *ServerItem {
	out := &ServerItem{
		Address:      in.Address,
		AppId:        in.AppId,
		Env:          in.Env,
		Hostname:     in.Hostname,
		RegTimestamp: in.RegTimestamp,
	}
	return out
}
