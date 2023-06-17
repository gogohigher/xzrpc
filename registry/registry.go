package registry

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	defaultPath    = "/_xzrpc_/registry"
	defaultTimeout = time.Minute * 5
)

// 注册中心

type XzRegistry struct {
	servers map[string]*ServerItem
	mu      sync.Mutex
	timeout time.Duration
}

type ServerItem struct {
	Addr  string
	start time.Time
}

func NewRegistry(timeout time.Duration) *XzRegistry {
	return &XzRegistry{
		servers: make(map[string]*ServerItem),
		timeout: timeout,
	}
}

// 添加服务实例
func (r *XzRegistry) putServer(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	server, ok := r.servers[addr]
	if !ok {
		r.servers[addr] = &ServerItem{
			Addr:  addr,
			start: time.Now(),
		}
	} else {
		// if exists, update start time to keep alive
		server.start = time.Now()
	}

	log.Println("xzrpc registry | register Server success, and addr: ", addr)

}

// 返回可用的服务列表
func (r *XzRegistry) getAliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	aliveServers := make([]string, 0)

	for addr, server := range r.servers {
		if r.timeout == 0 || server.start.Add(r.timeout).After(time.Now()) {
			aliveServers = append(aliveServers, addr)
		} else {
			delete(r.servers, addr)
		}
	}
	return aliveServers
}

// 简单实现
// GET：返回可用的服务列表
// POST：添加服务实例或者发送心跳
func (r *XzRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		w.Header().Set("X-Xzrpc-Servers", strings.Join(r.getAliveServers(), ","))
	case "POST":
		addr := req.Header.Get("X-Xzrpc-Server")
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.putServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *XzRegistry) HandleHTTP(registryPath string) {
	http.Handle(registryPath, r)
	log.Printf("xzrpc registry | registry path: %s\n", registryPath)
}
