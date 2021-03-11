package mware

import (
	logger "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"
	"net/http"
)

var log = logger.Logger("http/mware")

type MiddlewareOut struct {
	fx.Out

	Mware Middleware `group:"httpmiddleware"`
}

type Middleware func(h http.Handler) http.Handler
