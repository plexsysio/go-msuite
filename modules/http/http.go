package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/aloknerurkar/go-msuite/modules/diag/status"
	"github.com/aloknerurkar/go-msuite/modules/http/middleware"
	"github.com/aloknerurkar/go-msuite/utils"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	logger "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"
	nhttp "net/http"
)

var log = logger.Logger("http")

var Module = func(c config.Config) fx.Option {
	return fx.Options(
		utils.MaybeProvide(NewHTTPServerMux, c.IsSet("UseHTTP")),
		utils.MaybeProvide(NewGRPCGateway, c.IsSet("UseHTTP")),
		utils.MaybeProvide(mware.JWT, c.IsSet("UseJWT")),
		utils.MaybeProvide(mware.Tracing, c.IsSet("UseTracing")),
		utils.MaybeProvide(mware.Prometheus, c.IsSet("UsePrometheus")),
		utils.MaybeInvoke(NewHTTPServer, c.IsSet("UseHTTP")),
	)
}

type HTTPIn struct {
	fx.In

	Mux    *nhttp.ServeMux
	GRPC   *runtime.ServeMux
	Mwares []mware.Middleware `group:"httpmiddleware"`
}

func NewHTTPServerMux() *nhttp.ServeMux {
	return nhttp.NewServeMux()
}

func NewGRPCGateway() *runtime.ServeMux {
	return runtime.NewServeMux()
}

func NewHTTPServer(
	lc fx.Lifecycle,
	c config.Config,
	httpIn HTTPIn,
	st status.Manager,
) error {
	var httpPort int
	ok := c.Get("HTTPPort", &httpPort)
	if !ok {
		return errors.New("HTTP Port not provided")
	}
	if httpIn.GRPC != nil {
		httpIn.Mux.Handle("/", httpIn.GRPC)
	}
	var rootHandler nhttp.Handler = httpIn.Mux
	for _, v := range httpIn.Mwares {
		rootHandler = v(rootHandler)
	}
	httpServer := &nhttp.Server{Addr: fmt.Sprintf(":%d", httpPort), Handler: rootHandler}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Info("Starting http server")
				err := httpServer.ListenAndServe()
				if err != nil {
					log.Error("http server stopped ", err)
				}
			}()
			st.Report("HTTP Server", status.String(fmt.Sprintf("Running on port %d", httpPort)))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			defer st.Report("HTTP Server", status.String("Stopped"))
			return httpServer.Close()
		},
	})
	return nil
}
