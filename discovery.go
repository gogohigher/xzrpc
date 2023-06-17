package xzrpc

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

const (
	RandomStrategy     = iota // 随机策略
	RoundRobinStrategy        // 轮询
	// TODO 加权轮询、哈希一致性
)

type StrategyMode int

type Discovery interface {
	Refresh() error                            // 从注册中心更新服务列表
	Update(servers []string) error             // 手动更新服务列表
	Get(strategy StrategyMode) (string, error) // 根据负载均衡策略，选择一个服务实例
	GetAll() ([]string, error)                 // 返回所有的服务实例
}

// DefaultDiscovery 一个Discovery示例
type DefaultDiscovery struct {
	rand    *rand.Rand // random strategy 使用
	mu      sync.RWMutex
	servers []string // 服务集合
	index   int      // round-robin算法的起始下标，每次都随机产生一个，避免每次都从0开始
}

var _ Discovery = (*DefaultDiscovery)(nil)

func NewDefaultDiscovery(servers []string) *DefaultDiscovery {
	discovery := &DefaultDiscovery{
		rand:    rand.New(rand.NewSource(time.Now().Unix())),
		servers: servers,
	}
	discovery.index = discovery.rand.Intn(math.MaxInt64)
	return discovery
}

func (d *DefaultDiscovery) Refresh() error {
	return nil
}

func (d *DefaultDiscovery) Update(servers []string) error {
	d.mu.Lock()
	d.servers = servers
	d.mu.Unlock()
	return nil
}

// Get 根据策略，选择对应的服务
// 保证线程安全，感觉不需要整个方法加锁
func (d *DefaultDiscovery) Get(strategy StrategyMode) (string, error) {
	n := len(d.servers)
	switch strategy {
	case RandomStrategy:
		d.mu.RLock()
		defer d.mu.RUnlock()
		if n == 0 {
			return "", errors.New("xzrpc discovery | not found server")
		}
		k := d.rand.Intn(n)
		return d.servers[k], nil
	case RoundRobinStrategy:
		d.mu.Lock()
		defer d.mu.Unlock()
		if n == 0 {
			return "", errors.New("xzrpc discovery | not found server")
		}
		s := d.servers[d.index%n]
		d.index = (d.index + 1) % n
		return s, nil
	default:
		return "", fmt.Errorf("xzrpc discovery | not support %d strategy", strategy)
	}
}

// GetAll 返回所有的服务
func (d *DefaultDiscovery) GetAll() ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	servers := make([]string, len(d.servers))
	copy(servers, d.servers)
	return servers, nil
}
