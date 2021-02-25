package auth

import (
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/aloknerurkar/go-msuite/utils"
	"go.uber.org/fx"
)

var Module = func(c config.Config) fx.Option {
	return fx.Options(
		utils.MaybeProvide(NewJWTManager, c.IsSet("UseJWT")),
		utils.MaybeProvide(NewAclManager, c.IsSet("UseACL")),
	)
}
