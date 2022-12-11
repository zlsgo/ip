package ip

import (
	"context"
	"errors"
	"time"

	"github.com/sohaha/zlsgo/zhttp"
	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
)

var http = zutil.Once(func() *zhttp.Engine {
	h := zhttp.New()
	h.SetUserAgent(func() string {
		return ""
	})

	return h
})

// Generate 生成随机 IP
func Generate(start, end string) string {
	s, _ := znet.IPToLong(start)
	d, _ := znet.IPToLong(end)

	if (s == 0 && d == 0) || s >= d {
		return ""
	}

	return numToIp(int(s) + zstring.RandInt(0, int(d)-int(s)))
}

// NetWorkIP 获取外网 IP
func NetWorkIP() (ip string, err error) {
	h := http()
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()

	m := []netWorkIPFn{ipsb, ipapi, ipecho, ifconfig}
	c := make(chan string, len(m))
	for _, v := range m {
		go v(ctx, h, c)
	}

	select {
	case ip = <-c:
		cancel()
	case <-ctx.Done():
		err = errors.New("timeout")
	}

	return
}

type netWorkIPFn func(ctx context.Context, h *zhttp.Engine, c chan<- string)

func ipsb(ctx context.Context, h *zhttp.Engine, c chan<- string) {
	r, err := h.Get("https://api.ip.sb/ip", ctx)
	if err != nil {
		return
	}
	if r.StatusCode() != 200 {
		return
	}
	ip := zstring.TrimSpace(r.String())
	if ip == "" {
		return
	}
	c <- ip
}

func ipecho(ctx context.Context, h *zhttp.Engine, c chan<- string) {
	r, err := h.Get("https://ipecho.net/plain", ctx)
	if err != nil {
		return
	}
	if r.StatusCode() != 200 {
		return
	}
	ip := zstring.TrimSpace(r.String())
	if ip == "" {
		return
	}
	c <- ip
}

func ifconfig(ctx context.Context, h *zhttp.Engine, c chan<- string) {
	r, err := h.Get("https://ifconfig.me/ip", ctx)
	if err != nil {
		return
	}
	if r.StatusCode() != 200 {
		return
	}
	ip := zstring.TrimSpace(r.String())
	if ip == "" {
		return
	}
	c <- ip
}

func ipapi(ctx context.Context, h *zhttp.Engine, c chan<- string) {
	r, err := h.Get("http://ip-api.com/json/?lang=zh-CN", ctx)
	if err != nil {
		return
	}
	if r.StatusCode() != 200 {
		return
	}
	ip := r.JSON("query").String()
	if ip == "" {
		return
	}
	c <- ip
}
