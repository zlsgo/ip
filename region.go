package ip

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zhttp"
	"github.com/sohaha/zlsgo/zutil"
)

const dburl = "https://raw.githubusercontent.com/lionsoul2014/ip2region/master/data/ip2region.xdb"

// const globalRegionUrl = "https://raw.githubusercontent.com/lionsoul2014/ip2region/master/data/global_region.csv"

var db = zutil.Once(func() (searcher *Searcher) {
	dbPath := zfile.RealPath("ip.xdb")
	download := func() {
		downloadURL := dburl
		var r *zhttp.Res
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		r, err := http().Get("https://api.ip.sb/geoip", ctx)
		if err == nil {
			if r.JSON("country").String() == "China" {
				downloadURL = "https://ghproxy.com/" + downloadURL
			}
		} else {
			downloadURL = "https://ghproxy.com/" + downloadURL
		}
		r, err = http().Get(downloadURL)
		if err != nil {
			r, err = http().Get(dburl)
		}
		if err == nil {
			err = r.ToFile(dbPath)
		}
		zerror.Panic(err)
	}
	if zfile.FileSizeUint(dbPath) < 1024*1024*2 {
		download()
	}

	cBuff, err := loadContentFromFile(dbPath)
	if err != nil {
		_ = zfile.Remove(dbPath)
	}
	zerror.Panic(err)

	searcher, err = NewWithBuffer(cBuff)
	zerror.Panic(err)
	return
})

type Res struct {
	Country  string `json:"country"`
	City     string `json:"city"`
	Province string `json:"province"`
}

// Region 获取 IP 地理位置
func Region(ip string) (r Res, err error) {
	searcher := db()
	if searcher == nil {
		return r, errors.New("数据库初始化失败, 请重试")
	}
	s, err := searcher.SearchByStr(ip)
	if err != nil {
		return r, err
	}
	res := strings.Split(s, "|")
	if len(res) < 5 {
		return r, errors.New("IP 解析失败")
	}

	r.Country = zutil.IfVal(res[0] == "0", "", res[0]).(string)
	r.Province = zutil.IfVal(res[2] == "0", "", res[2]).(string)
	deCity := r.Province
	if deCity == "" {
		deCity = r.Country
	}
	r.City = zutil.IfVal(res[3] == "0", deCity, res[3]).(string)
	return r, nil
}
