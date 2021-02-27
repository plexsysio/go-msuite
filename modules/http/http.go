package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/aloknerurkar/go-msuite/utils"
	logger "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"
	"net/http"
)

var log = logger.Logger("http")

var Module = func(c config.Config) fx.Option {
	return fx.Options(
		utils.MaybeProvide(NewHTTPServerMux, c.IsSet("UseHTTP")),
		utils.MaybeInvoke(NewHTTPServer, c.IsSet("UseHTTP")),
	)
}

func NewHTTPServerMux() *http.ServeMux {
	return http.NewServeMux()
}

func NewHTTPServer(lc fx.Lifecycle, c config.Config, httpMux *http.ServeMux) error {
	var httpPort int
	ok := c.Get("HTTPPort", &httpPort)
	if !ok {
		return errors.New("HTTP Port not provided")
	}
	httpServer := &http.Server{Addr: fmt.Sprintf(":%d", httpPort), Handler: httpMux}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Info("Starting http server")
				err := httpServer.ListenAndServe()
				if err != nil {
					log.Error("http server stopped ", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return httpServer.Close()
		},
	})
	return nil
}
