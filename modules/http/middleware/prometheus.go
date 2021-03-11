package mware

import (
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
)

func Prometheus() MiddlewareOut {
	mdlw := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})
	return MiddlewareOut{
		Mware: std.HandlerProvider("", mdlw),
	}
}
