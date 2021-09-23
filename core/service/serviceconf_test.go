package service

import (
	"testing"

	"github.com/tal-tech/go-zero/core/logx"
)

func TestServiceConf(t *testing.T) {
	c := ServiceConf{
		Name: "foo",
		Log: logx.LogConf{
			Mode: "console",
		},
		Mode:          "dev",
		SlowThreshold: 500,
	}
	c.MustSetUp()
}
