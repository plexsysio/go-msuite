package grpcServer

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	logger "github.com/ipfs/go-log/v2"
	"github.com/plexsysio/go-msuite/modules/config"
	"github.com/plexsysio/go-msuite/modules/diag/status"
	"github.com/plexsysio/go-msuite/modules/grpc/middleware"
	"github.com/plexsysio/go-msuite/modules/grpc/transport/mux"
	"github.com/plexsysio/go-msuite/modules/grpc/transport/p2pgrpc"
	"github.com/plexsysio/go-msuite/modules/grpc/transport/tcp"
	"github.com/plexsysio/go-msuite/utils"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

var log = logger.Logger("grpc_service")

type GrpcServerParams struct {
	fx.In

	Opts      []grpc.ServerOption
	Listnr    *grpcmux.Mux
	StManager status.Manager `optional:"true"`
}

func New(
	lc fx.Lifecycle,
	params GrpcServerParams,
) (*grpc.Server, error) {
	rpcSrv := grpc.NewServer(params.Opts...)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Info("Starting GRPC server")
				err := rpcSrv.Serve(params.Listnr)
				if err != nil {
					log.Error("Failed to serve gRPC", err.Error())
					if params.StManager != nil {
						params.StManager.Report("GRPC server",
							status.String(fmt.Sprintf("Failed Err:%s", err.Error())))

					}
				}
			}()
			if params.StManager != nil {
				params.StManager.Report("GRPC server", status.String("Running"))
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if params.StManager != nil {
				defer params.StManager.Report("GRPC server", status.String("Stopped"))
			}
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
	return fx.Options(
		utils.MaybeProvide(tcp.NewTCPListener, c.IsSet("UseTCP")),
		utils.MaybeProvide(p2pgrpc.NewP2PListener, c.IsSet("UseP2P")),
	)
}

func Middleware(c config.Config) fx.Option {
	return fx.Options(
		utils.MaybeOption(mware.JwtAuth, c.IsSet("UseJWT")),
		utils.MaybeOption(mware.TracerModule, c.IsSet("UseTracing")),
		utils.MaybeOption(mware.Prometheus, c.IsSet("UsePrometheus")),
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
