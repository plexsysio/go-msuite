package http

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
	"go.uber.org/fx"
)

type MiddlewareOut struct {
	fx.Out

	Mware Middleware `group:"httpmiddleware"`
}

type Middleware func(h http.Handler) http.Handler

func CORS() MiddlewareOut {
	return MiddlewareOut{
		Mware: cors.Default().Handler,
	}
}

func JWT() MiddlewareOut {
	return MiddlewareOut{
		Mware: func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Info("JWT called")
				next.ServeHTTP(w, r)
			})
		},
	}
}

func Tracing() MiddlewareOut {
	return MiddlewareOut{
		Mware: func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Infof("Tracing called")
				next.ServeHTTP(w, r)
			})
		},
	}
}

var Prometheus = fx.Options(
	fx.Provide(PromMware),
	fx.Invoke(Register),
)

func PromMware() MiddlewareOut {
	mdlw := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})
	return MiddlewareOut{
		Mware: std.HandlerProvider("", mdlw),
	}
}

func Register(mux *http.ServeMux, reg *prometheus.Registry) {
	mux.Handle("/metrics", promhttp.InstrumentMetricHandler(
		reg,
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
	))
}
