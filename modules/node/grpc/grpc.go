package grpcsvc

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	logger "github.com/ipfs/go-log/v2"
	"github.com/plexsysio/go-msuite/modules/config"
	"github.com/plexsysio/go-msuite/modules/diag/status"
	"github.com/plexsysio/go-msuite/modules/grpc/client"
	"github.com/plexsysio/go-msuite/modules/grpc/mux"
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

func OptsAggregator(params ServerOptsParams) (outOpts []grpc.ServerOption) {
	outOpts = append(
		outOpts,
		grpc_middleware.WithUnaryServerChain(params.UnaryOpts...),
		grpc_middleware.WithStreamServerChain(params.StreamOpts...),
	)
	return outOpts
}

func Transport(c config.Config) fx.Option {
	return fx.Options(
		fx.Provide(NewMuxedListener),
		utils.MaybeProvide(NewTCPListener, c.IsSet("UseTCP")),
		utils.MaybeProvide(NewP2PListener, c.IsSet("UseP2P")),
	)
}

func Middleware(c config.Config) fx.Option {
	return fx.Options(
		utils.MaybeOption(JwtAuth, c.IsSet("UseJWT")),
		utils.MaybeOption(TracerModule, c.IsSet("UseTracing")),
		utils.MaybeOption(Prometheus, c.IsSet("UsePrometheus")),
	)
}

func Client(c config.Config) fx.Option {
	return fx.Options(
		utils.MaybeProvide(grpcclient.NewP2PClientService, c.IsSet("UseP2P")),
		utils.MaybeProvide(grpcclient.NewStaticClientService, !c.IsSet("UseP2P") && c.IsSet("UseStaticDiscovery")),
	)
}

var Module = func(c config.Config) fx.Option {
	return fx.Options(
		Transport(c),
		Middleware(c),
		Client(c),
		fx.Provide(OptsAggregator),
		fx.Provide(New),
	)
}
