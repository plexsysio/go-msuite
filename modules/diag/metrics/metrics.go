package metrics

import (
	"context"
	"errors"

	gtrace "github.com/moxiaomomo/grpc-jaeger"
	"github.com/opentracing/opentracing-go"
	"github.com/plexsysio/go-msuite/modules/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.uber.org/fx"
)

func New() *prometheus.Registry {
	r := prometheus.NewRegistry()

	r.MustRegister(
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
			Namespace: "msuite",
		}),
		collectors.NewGoCollector(),
	)

	return r
}

func NewTracer(lc fx.Lifecycle, conf config.Config) (opentracing.Tracer, error) {
	svcName := "default"
	conf.Get("TracingName", &svcName)

	var tHost string
	ok := conf.Get("TracingHost", &tHost)
	if !ok {
		return nil, errors.New("Tracing host not specified")
	}

	tracer, closer, err := gtrace.NewJaegerTracer(svcName, tHost)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return closer.Close()
		},
	})

	return tracer, nil
}
