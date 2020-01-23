package grpcServer

import (
	"context"
	"errors"
	"github.com/aloknerurkar/go-msuite/modules/config"
	gtrace "github.com/moxiaomomo/grpc-jaeger"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

var TracerModule = fx.Options(
	fx.Provide(JaegerTracerOptions),
)

type TracerOpts struct {
	fx.Out

	UOut grpc.UnaryServerInterceptor `group:"unary_opts"`
}

func JaegerTracerOptions(lc fx.Lifecycle,
	conf config.Config) (params TracerOpts, retErr error) {

	useTracing, ok := conf.Get("use_tracing").(bool)
	if !ok {
		useTracing = false
	}
	if useTracing {
		svcName, ok := conf.Get("service_name").(string)
		if !ok {
			svcName = "default"
		}
		tHost, ok := conf.Get("tracing_host").(string)
		if !ok {
			retErr = errors.New("Tracing host not specified")
			return
		}
		tracer, closer, err := gtrace.NewJaegerTracer(svcName, tHost)
		if err != nil {
			retErr = err
			return
		}
		lc.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				return closer.Close()
			},
		})
		log.Info("Registering Jaeger tracer options")
		params.UOut = gtrace.ServerInterceptor(tracer)
		return
	}
	return
}
