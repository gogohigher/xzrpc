package main

import (
	_const "github.com/gogohigher/xzrpc/pkg/const"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestRegister(t *testing.T) {

	data := url.Values{}
	data.Set("env", _const.TEST_ENV)
	data.Set("appid", _const.TEST_APPID)
	data.Set("hostname", _const.TEST_HOST_NAME)
	data.Set("address", "tcp@127.0.0.1:8081")

	client := &http.Client{}
	registry := "http://127.0.0.1:9999/_xzrpc_/registry"
	request, _ := http.NewRequest("POST", registry, strings.NewReader(data.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, err := client.Do(request)
	if err != nil {
		log.Println("xzrpc server | failed to send heartbeat: ", err)
	}
	log.Println("TestRegister success")
}

func TestGet(t *testing.T) {

	data := url.Values{}
	data.Set("env", _const.TEST_ENV)
	data.Set("appid", _const.TEST_APPID)

	client := &http.Client{}
	registry := "http://127.0.0.1:9999/_xzrpc_/get"
	request, _ := http.NewRequest("POST", registry, strings.NewReader(data.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(request)
	if err != nil {
		log.Println("xzrpc server | failed to send heartbeat: ", err)
		return
	}
	servers := resp.Header.Get("X-Xzrpc-Servers")
	parts := strings.Split(servers, ",")

	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			log.Println("TestGet 获取到注册的地址: ", part)
		}
	}

	log.Println("TestGet success")
}
