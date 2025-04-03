package ip

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/pion/stun"
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
	ip, err = stunIP()
	if err == nil {
		return
	}

	h := http()
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()

	m := []netWorkIPFn{ipsb, ipapi, ipecho, ifconfigCo, ifconfigMe, ipzcorky, ipinfo}
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

func ipzcorky(ctx context.Context, h *zhttp.Engine, c chan<- string) {
	r, err := h.Get("https://ip.zcorky.com", ctx)
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

func ipinfo(ctx context.Context, h *zhttp.Engine, c chan<- string) {
	r, err := h.Get("https://ipinfo.io/json", ctx)
	if err != nil {
		return
	}
	if r.StatusCode() != 200 {
		return
	}

	ip := zstring.TrimSpace(r.JSON("ip").String())
	if ip == "" {
		return
	}
	c <- ip
}

func ifconfigMe(ctx context.Context, h *zhttp.Engine, c chan<- string) {
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

func ifconfigCo(ctx context.Context, h *zhttp.Engine, c chan<- string) {
	r, err := h.Get("https://ifconfig.co/ip", ctx)
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

func stunIP() (ip string, err error) {
	addrs := []string{"stun.chat.bilibili.com:3478", "stun.cloudflare.com:3478", "stun.l.google.com:19302"}
	addrsSize := len(addrs)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()

	ipChan := make(chan string, addrsSize)

	for i := range addrs {
		addr := addrs[i]
		go func() {
			var ip string

			defer func() {
				ipChan <- ip
			}()

			var d net.Dialer
			conn, err := d.DialContext(ctx, "udp4", addr)
			if err != nil {
				return
			}

			c, err := stun.NewClient(conn)
			if err != nil {
				return
			}

			_ = c.Do(stun.MustBuild(stun.TransactionID, stun.BindingRequest), func(res stun.Event) {
				if res.Error == nil {
					var xorAddr stun.XORMappedAddress
					if getErr := xorAddr.GetFrom(res.Message); getErr != nil {
						return
					}
					ip = xorAddr.IP.String()
				}
			})
			_ = c.Close()
		}()
	}

g:
	for {
		select {
		case ip = <-ipChan:
			addrsSize--
			if ip != "" || addrsSize == 0 {
				cancel()
				break g
			}
		case <-ctx.Done():
			break g
		}
	}

	if ip == "" {
		return "", errors.New("unable to connect to stun server")
	}

	return ip, nil
}
