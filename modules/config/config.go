package config

import (
	"github.com/aloknerurkar/go-msuite/modules/config/json"
	"go.uber.org/fx"
)

type Config interface {
	Get(string) interface{}
}

func NewConf() Config {
	return jsonConf.DefaultConfig()
}

var Module = fx.Options(
	fx.Provide(NewConf),
)
