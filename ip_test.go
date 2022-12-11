package ip

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/znet"
)

func TestNetWorkIP(t *testing.T) {
	tt := zlsgo.NewTest(t)
	ip, err := NetWorkIP()
	tt.NoError(err)
	t.Log(ip)
}

func TestGenerate(t *testing.T) {
	t.Log(znet.IPToLong("36.56.0.0"))
	t.Log(Generate("36.56.0.0", "36.63.255.255"))
	t.Log(Generate("36.56.0.0", "36.63.255.255"))
	t.Log(Generate("36.56.0.0", "36.6.255.255"))
	t.Log(Region(Generate("36.56.0.0", "36.63.255.255")))
	t.Log(Region(Generate("116.23.17.0", "116.23.171.255")))
	t.Log(Region(Generate("116.23.17.0", "116.23.171.255")))
	t.Log(Region(Generate("116.23.171.0", "116.23.171.255")))
}
