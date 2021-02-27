package utils

import "go.uber.org/fx"

func MaybeProvide(opt interface{}, enable bool) fx.Option {
	if enable {
		return fx.Provide(opt)
	}
	return fx.Options()
}

func MaybeInvoke(opt interface{}, enable bool) fx.Option {
	if enable {
		return fx.Invoke(opt)
	}
	return fx.Options()
}

func MaybeOption(opt fx.Option, enable bool) fx.Option {
	if enable {
		return opt
	}
	return fx.Options()
}
