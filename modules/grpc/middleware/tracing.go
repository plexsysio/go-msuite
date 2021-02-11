package mware

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

func JaegerTracerOptions(
	lc fx.Lifecycle,
	conf config.Config,
) (params TracerOpts, retErr error) {
	svcName := "default"
	conf.Get("TracingName", &svcName)
	var tHost string
	ok := conf.Get("TracingHost", &tHost)
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
	// log.Info("Registering Jaeger tracer options")
	params.UOut = gtrace.ServerInterceptor(tracer)
	return
}
