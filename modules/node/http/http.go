package http

import (
	"context"
	"errors"
	"fmt"
	nhttp "net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	logger "github.com/ipfs/go-log/v2"
	"github.com/plexsysio/go-msuite/modules/config"
	"github.com/plexsysio/go-msuite/modules/diag/status"
	"github.com/plexsysio/go-msuite/utils"
	"go.uber.org/fx"
)

var log = logger.Logger("http")

var Module = func(c config.Config) fx.Option {
	return fx.Options(
		fx.Provide(NewHTTPServerMux),
		fx.Provide(NewGRPCGateway),
		fx.Invoke(NewHTTPServer),
		utils.MaybeProvide(JWT, c.IsSet("UseAuth")),
		utils.MaybeProvide(Tracing, c.IsSet("UseTracing")),
		utils.MaybeOption(Prometheus, c.IsSet("UsePrometheus")),
		utils.MaybeInvoke(RegisterDebug, c.IsSet("UseDebug")),
	)
}

type HTTPIn struct {
	fx.In

	Mux    *nhttp.ServeMux
	GRPC   *runtime.ServeMux
	Mwares []Middleware `group:"httpmiddleware"`
}

func NewHTTPServerMux() *nhttp.ServeMux {
	return nhttp.NewServeMux()
}

func NewGRPCGateway() *runtime.ServeMux {
	return runtime.NewServeMux()
}

type httpReporter struct {
	port    int
	stopped chan struct{}
}

func (h *httpReporter) Status() interface{} {
	select {
	case <-h.stopped:
		return "stopped"
	default:
	}
	return fmt.Sprintf("running on port %d", h.port)
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
			started, stopped := make(chan struct{}), make(chan struct{})
			go func() {
				close(started)
				defer close(stopped)

				log.Info("Starting http server")
				err := httpServer.ListenAndServe()
				if err != nil {
					log.Error("http server stopped ", err)
				}
			}()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-started:
			}
			st.AddReporter("HTTP Server", &httpReporter{port: httpPort, stopped: stopped})
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return httpServer.Shutdown(ctx)
		},
	})
	return nil
}
