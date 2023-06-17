package registry

import (
	"log"
	"net/http"
	"time"
)

// 发送心跳给注册中心
// TODO 注册中心是否需要集群？这里是不是也可以支持一下
func sendHeartbeat(registry, addr string) error {
	log.Printf("xzrpc registry | send heartbeat to %s address\n", addr)
	client := &http.Client{}

	request, _ := http.NewRequest("POST", registry, nil)
	request.Header.Set("X-Xzrpc-Server", addr)

	_, err := client.Do(request)
	if err != nil {
		log.Println("xzrpc server | failed to send heartbeat: ", err)
	}
	return err
}

// SendHeartbeat 每隔一段时间发送心跳
// 发送心跳的间隔时间应该小于注册中心将服务移除的时间
func SendHeartbeat(registry, addr string, duration time.Duration) {
	if duration == 0 {
		duration = defaultTimeout - time.Minute*time.Duration(1)
	}
	err := sendHeartbeat(registry, addr)
	t := time.NewTicker(duration)

	defer t.Stop()

	go func() {
		for {
			select {
			case <-t.C:
				err = sendHeartbeat(registry, addr)
				if err != nil {
					return
				}
			}
		}
	}()
}
