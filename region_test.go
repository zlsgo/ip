package ip

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zfile"
)

func TestRegion(t *testing.T) {
	defer zfile.Remove("ip.xdb")
	tt := zlsgo.NewTest(t)

	r, err := Region("180.149.130.16")

	t.Log(r)
	tt.NoError(err)
	tt.Equal("中国", r.Country)
	tt.Equal("北京市", r.City)
}
