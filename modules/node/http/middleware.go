package http

import (
	"expvar"
	"net/http"
	"net/http/pprof"

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

func RegisterDebug(mux *http.ServeMux) {
	mux.Handle("/debug/pprof", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := r.URL
		u.Path += "/"
		http.Redirect(w, r, u.String(), http.StatusPermanentRedirect)
	}))

	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))

	mux.Handle("/debug/vars", expvar.Handler())
}
