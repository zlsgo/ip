package ip

import (
	"testing"
)

func TestGetGlobalCountry(t *testing.T) {
	t.Log(len(GetGlobalCountry()))
}
