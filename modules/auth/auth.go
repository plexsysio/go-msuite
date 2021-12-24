package auth

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewJWTManager),
	fx.Provide(
		fx.Annotate(
			NewAclManager,
			fx.ParamTags(``, `optional:"true"`),
		),
	),
)
