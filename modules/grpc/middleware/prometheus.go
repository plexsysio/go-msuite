package mware

import (
	"github.com/plexsysio/go-msuite/modules/config"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

var Prometheus = fx.Options(
	fx.Provide(PromOptions),
	fx.Invoke(PromRegister),
)

type PrometheusOpts struct {
	fx.Out

	UOut grpc.UnaryServerInterceptor  `group:"unary_opts"`
	SOut grpc.StreamServerInterceptor `group:"stream_opts"`
}

func PromOptions() (params PrometheusOpts, err error) {
	params.SOut = grpc_prometheus.StreamServerInterceptor
	params.UOut = grpc_prometheus.UnaryServerInterceptor
	return
}

func PromRegister(c config.Config, s *grpc.Server) {
	grpc_prometheus.Register(s)
	if c.IsSet("UsePrometheusLatency") {
		grpc_prometheus.EnableHandlingTimeHistogram()
	}
}
