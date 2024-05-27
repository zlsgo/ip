package main

import (
	"github.com/sohaha/zlsgo/zlog"
	"github.com/zlsgo/ip"
)

func main() {
	// 查询 IP 信息
	r, err := ip.Region("159.75.67.22")
	if err != nil {
		zlog.Error(err)
		return
	}

	zlog.Debug(r)
	// r.Country 中国
	// r.City 北京市

	// 查询外网 IP
	zlog.Debug(ip.NetWorkIP())
}
