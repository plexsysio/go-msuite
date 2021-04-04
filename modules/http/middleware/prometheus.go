package mware

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
	"go.uber.org/fx"
	"net/http"
)

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

func Register(mux *http.ServeMux) {
	mux.Handle("/v1/metrics", promhttp.Handler())
}
