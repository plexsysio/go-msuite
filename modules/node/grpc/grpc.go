package grpcsvc

import (
	"context"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	logger "github.com/ipfs/go-log/v2"
	"github.com/plexsysio/go-msuite/modules/config"
	"github.com/plexsysio/go-msuite/modules/diag/status"
	grpcclient "github.com/plexsysio/go-msuite/modules/grpc/client"
	grpcmux "github.com/plexsysio/go-msuite/modules/grpc/mux"
	"github.com/plexsysio/go-msuite/utils"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

var log = logger.Logger("grpc_service")

type GrpcServerParams struct {
	fx.In

	Opts      []grpc.ServerOption
	Listnr    *grpcmux.Mux
	StManager status.Manager
}

type grpcReporter struct {
	stopped chan struct{}
}

func (s *grpcReporter) Status() interface{} {
	select {
	case <-s.stopped:
		return "stopped"
	default:
	}
	return "running"
}

func New(
	lc fx.Lifecycle,
	params GrpcServerParams,
) (*grpc.Server, error) {
	rpcSrv := grpc.NewServer(params.Opts...)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			stopped := make(chan struct{})
			go func() {
				defer close(stopped)
				log.Info("Starting GRPC server")
				err := rpcSrv.Serve(params.Listnr)
				if err != nil {
					log.Error("Failed to serve gRPC", err.Error())
				}
			}()
			params.StManager.AddReporter("GRPC Server", &grpcReporter{stopped: stopped})
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
		utils.MaybeInvoke(grpcclient.NewP2PClientAdvertiser, c.IsSet("UseP2P")),
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
