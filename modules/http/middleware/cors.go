package mware

import (
	"github.com/rs/cors"
)

func CORS() MiddlewareOut {
	return MiddlewareOut{
		Mware: cors.Default().Handler,
	}
}
