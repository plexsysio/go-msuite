package grpcServer

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	logger "github.com/ipfs/go-log"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"net"
)

var log = logger.Logger("grpc_service")

type GrpcServerParams struct {
	fx.In

	Opts   []grpc.ServerOption
	Listnr net.Listener
}

func New(lc fx.Lifecycle, params GrpcServerParams) (*grpc.Server, error) {

	rpcSrv := grpc.NewServer(params.Opts...)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Starting GRPC server")
			return rpcSrv.Serve(params.Listnr)
		},
		OnStop: func(ctx context.Context) error {
			rpcSrv.Stop()
			return nil
		},
	})
	return rpcSrv, nil
}

type ServerOptsParams struct {
	fx.In

	UnaryOpts  []grpc.UnaryServerInterceptor  `group:"unary_opts"`
	StreamOpts []grpc.StreamServerInterceptor `group:"stream_opts"`
}

func OptsAggregator(params ServerOptsParams) []grpc.ServerOption {

	outOpts := make([]grpc.ServerOption, 0)
	outOpts = append(outOpts, grpc_middleware.WithUnaryServerChain(params.UnaryOpts...))
	outOpts = append(outOpts, grpc_middleware.WithStreamServerChain(params.StreamOpts...))
	return outOpts
}

var Module = fx.Options(
	fx.Provide(OptsAggregator),
	fx.Invoke(New),
)
