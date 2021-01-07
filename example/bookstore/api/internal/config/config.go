package config

import (
	"github.com/3Rivers/go-zero/rest"
	"github.com/3Rivers/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Add   zrpc.RpcClientConf
	Check zrpc.RpcClientConf
}
