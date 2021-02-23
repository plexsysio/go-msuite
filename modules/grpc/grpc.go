package grpcServer

import (
	"context"
	"fmt"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/aloknerurkar/go-msuite/modules/grpc/middleware"
	"github.com/aloknerurkar/go-msuite/modules/grpc/transport/mux"
	"github.com/aloknerurkar/go-msuite/modules/grpc/transport/p2pgrpc"
	"github.com/aloknerurkar/go-msuite/modules/grpc/transport/tcp"
	"github.com/aloknerurkar/go-msuite/utils"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	logger "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

var log = logger.Logger("grpc_service")

type GrpcServerParams struct {
	fx.In

	Opts   []grpc.ServerOption
	Listnr *grpcmux.Mux
}

func New(lc fx.Lifecycle, params GrpcServerParams) (*grpc.Server, error) {
	rpcSrv := grpc.NewServer(params.Opts...)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Info("Starting GRPC server")
				err := rpcSrv.Serve(params.Listnr)
				if err != nil {
					log.Error("Failed to serve gRPC", err.Error())
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping GRPC server")
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
	log.Info("Server opts:", params)
	outOpts := make([]grpc.ServerOption, 0)
	outOpts = append(outOpts, grpc_middleware.WithUnaryServerChain(params.UnaryOpts...))
	outOpts = append(outOpts, grpc_middleware.WithStreamServerChain(params.StreamOpts...))
	return outOpts
}

func Transport(c config.Config) fx.Option {
	fmt.Println("Transport")
	return fx.Options(
		utils.MaybeProvide(tcp.NewTCPListener, c.IsSet("UseTCP")),
		utils.MaybeProvide(p2pgrpc.NewP2PListener, c.IsSet("UseP2P")),
	)
}

func Middleware(c config.Config) fx.Option {
	fmt.Println("Middleware")
	return fx.Options(
		utils.MaybeProvide(mware.JwtAuth, c.IsSet("UseJWT")),
		utils.MaybeProvide(mware.TracerModule, c.IsSet("UseTracing")),
	)
}

var Module = func(c config.Config) fx.Option {
	return fx.Options(
		Transport(c),
		Middleware(c),
		grpcmux.Module,
		fx.Provide(OptsAggregator),
		fx.Provide(New),
	)
}
