package xzrpc

import (
	_const "github.com/gogohigher/xzrpc/pkg/const"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultUpdateTimeout = time.Second * 10

type RegistryDiscovery struct {
	*DefaultDiscovery
	// 注册中心地址
	registry string
	// 服务列表过期时间
	timeout time.Duration
	// 最后从注册中心更新服务列表的时间，超过过期时间，就需要请求注册中心，拿到最新的服务列表
	// TODO 这里的过期时间很难确定，更好的方式是加上注册中心如果收到服务的变化，应该通知客户端，这个功能后续加上
	lastUpdate time.Time
}

func NewRegistryDiscovery(registryAddr string, timeout time.Duration) *RegistryDiscovery {
	rd := &RegistryDiscovery{
		DefaultDiscovery: NewDefaultDiscovery(make([]string, 0)),
		registry:         registryAddr,
		timeout:          timeout,
	}
	return rd
}

// TODO 是不是应该也支持一个？好像没必要！
func (rd *RegistryDiscovery) Update(servers []string) error {
	rd.mu.Lock()
	defer rd.mu.Unlock()
	rd.servers = servers
	rd.lastUpdate = time.Now()
	return nil
}

func (rd *RegistryDiscovery) Refresh() error {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	if rd.lastUpdate.Add(rd.timeout).After(time.Now()) {
		// 没有到时间，不需要更新
		return nil
	}
	//resp, err := http.Get(rd.registry)

	data := url.Values{}
	data.Set("env", _const.TEST_ENV)
	data.Set("appid", _const.TEST_APPID)
	client := &http.Client{}
	registry := "http://127.0.0.1:9999/_xzrpc_/get" // TODO 这个是不是应该封装一下
	request, _ := http.NewRequest("POST", registry, strings.NewReader(data.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(request)

	if err != nil {
		log.Println("xzrpc registry | refresh err: ", err)
		return err
	}
	servers := resp.Header.Get("X-Xzrpc-Servers")
	parts := strings.Split(servers, ",")

	rd.servers = make([]string, 0, len(parts))

	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			rd.servers = append(rd.servers, part)
		}
	}
	rd.lastUpdate = time.Now()
	return nil
}

func (rd *RegistryDiscovery) Get(strategy StrategyMode) (string, error) {
	//  1. 判断服务列表是否过期
	if err := rd.Refresh(); err != nil {
		return "", err
	}
	return rd.DefaultDiscovery.Get(strategy)
}

func (rd *RegistryDiscovery) GetAll() ([]string, error) {
	if err := rd.Refresh(); err != nil {
		return nil, err
	}
	return rd.DefaultDiscovery.GetAll()
}
