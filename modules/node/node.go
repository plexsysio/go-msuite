package node

import (
	"context"
	"errors"
	"github.com/SWRMLabs/ss-store"
	ipfslite "github.com/hsanjuan/ipfs-lite"
	ds "github.com/ipfs/go-datastore"
	logger "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/routing"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/plexsysio/dLocker"
	"github.com/plexsysio/go-msuite/core"
	"github.com/plexsysio/go-msuite/modules/auth"
	"github.com/plexsysio/go-msuite/modules/config"
	"github.com/plexsysio/go-msuite/modules/diag/status"
	"github.com/plexsysio/go-msuite/modules/events"
	"github.com/plexsysio/go-msuite/modules/grpc/client"
	"github.com/plexsysio/go-msuite/modules/node/grpc"
	mhttp "github.com/plexsysio/go-msuite/modules/node/http"
	"github.com/plexsysio/go-msuite/modules/node/ipfs"
	"github.com/plexsysio/go-msuite/modules/node/locker"
	"github.com/plexsysio/go-msuite/modules/repo"
	"github.com/plexsysio/go-msuite/modules/repo/fsrepo"
	"github.com/plexsysio/go-msuite/modules/sharedStorage"
	"github.com/plexsysio/go-msuite/utils"
	"github.com/plexsysio/taskmanager"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"net/http"
	"time"
)

type FxLog struct{}

var log = logger.Logger("node")

func (f *FxLog) Printf(msg string, args ...interface{}) {
	log.Infof(msg, args...)
}

var authModule = func(c config.Config) fx.Option {
	return fx.Options(
		utils.MaybeProvide(auth.NewJWTManager, c.IsSet("UseJWT")),
		utils.MaybeProvide(auth.NewAclManager, c.IsSet("UseACL")),
	)
}

func New(bCfg config.Config) (core.Service, error) {
	r, err := fsrepo.CreateOrOpen(bCfg)
	if err != nil {
		return nil, err
	}
	svc := &impl{}

	var svcName string
	nameFound := bCfg.Get("ServiceName", &svcName)
	if !nameFound {
		return nil, errors.New("service name not configured")
	}

	var tmCount int
	found := bCfg.Get("TMWorkersMin", &tmCount)

	app := fx.New(
		fx.Logger(&FxLog{}),
		fx.Provide(func() (context.Context, context.CancelFunc) {
			return context.WithCancel(context.Background())
		}),
		fx.Provide(func(lc fx.Lifecycle) (repo.Repo, config.Config, ds.Batching) {
			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					log.Debugf("closing repo")
					defer log.Debugf("closed repo")
					return r.Close()
				},
			})
			return r, r.Config(), r.Datastore()
		}),
		utils.MaybeProvide(func(ctx context.Context, lc fx.Lifecycle) *taskmanager.TaskManager {
			tm := taskmanager.New(tmCount, tmCount*3, time.Second*15)
			lc.Append(fx.Hook{
				OnStop: func(c context.Context) error {
					log.Debugf("stopping taskmanager")
					defer log.Debugf("stopped taskmanager")
					tm.Stop()
					return nil
				},
			})
			return tm
		}, found),
		fx.Provide(func() string {
			return svcName
		}),
		utils.MaybeOption(fx.Provide(status.New), bCfg.IsSet("UseHTTP")),
		utils.MaybeOption(locker.Module, bCfg.IsSet("UseLocker")),
		authModule(r.Config()),
		utils.MaybeOption(ipfs.Module, bCfg.IsSet("UseP2P")),
		utils.MaybeOption(grpcsvc.Module(r.Config()), bCfg.IsSet("UseGRPC")),
		mhttp.Module(r.Config()),
		utils.MaybeOption(fx.Provide(events.NewEventsSvc), bCfg.IsSet("UseP2P")),
		utils.MaybeOption(fx.Provide(sharedStorage.NewSharedStoreProvider), bCfg.IsSet("UseP2P")),
		fx.Invoke(func(lc fx.Lifecycle, cancel context.CancelFunc) {
			lc.Append(fx.Hook{
				OnStop: func(c context.Context) error {
					cancel()
					return nil
				},
			})
		}),
		fx.Populate(svc),
	)

	svc.App = app
	return svc, nil
}

type impl struct {
	fx.In

	*fx.App `optional:"true"`
	Ctx     context.Context
	Cancel  context.CancelFunc

	R    repo.Repo
	Am   auth.ACL                 `optional:"true"`
	Tm   *taskmanager.TaskManager `optional:"true"`
	Lk   dLocker.DLocker          `optional:"true"`
	Rsrv *grpc.Server             `optional:"true"`
	Mx   *http.ServeMux           `optional:"true"`
	H    host.Host                `optional:"true"`
	Dht  routing.Routing          `optional:"true"`
	P    *ipfslite.Peer           `optional:"true"`
	Ps   *pubsub.PubSub           `optional:"true"`
	Disc discovery.Discovery      `optional:"true"`
	St   store.Store              `optional:"true"`
	Jm   auth.JWTManager          `optional:"true"`
	Ev   events.Events            `optional:"true"`
	Cs   grpcclient.ClientSvc     `optional:"true"`
	ShSt sharedStorage.Provider   `optional:"true"`
}

func (s *impl) Repo() repo.Repo {
	return s.R
}

func (s *impl) TM() (*taskmanager.TaskManager, error) {
	if s.Tm == nil {
		return nil, errors.New("Taskmanager not configured")
	}
	return s.Tm, nil
}

func (s *impl) Node() (core.Node, error) {
	if s.H == nil {
		return nil, errors.New("Node not configured")
	}
	return s, nil
}

// P2P API
func (s *impl) P2P() core.P2P {
	return s
}

func (s *impl) Host() host.Host {
	return s.H
}

func (s *impl) Routing() routing.Routing {
	return s.Dht
}

func (s *impl) Discovery() discovery.Discovery {
	return s.Disc
}

// Pubsub API
func (s *impl) Pubsub() *pubsub.PubSub {
	return s.Ps
}

// IPFS API
func (s *impl) IPFS() *ipfslite.Peer {
	return s.P
}

// Auth API
func (s *impl) Auth() core.Auth {
	return s
}

func (s *impl) JWT() (auth.JWTManager, error) {
	if s.Jm == nil {
		return nil, errors.New("JWT not configured")
	}
	return s.Jm, nil
}

func (s *impl) ACL() (auth.ACL, error) {
	if s.Am == nil {
		return nil, errors.New("ACL manager not configured")
	}
	return s.Am, nil
}

func (s *impl) GRPC() (core.GRPC, error) {
	if s.Rsrv == nil {
		return nil, errors.New("GRPC service not configured")
	}
	return s, nil
}

func (s *impl) Server() *grpc.Server {
	return s.Rsrv
}

func (s *impl) Client(ctx context.Context, name string) (*grpc.ClientConn, error) {
	if s.Cs == nil {
		return nil, errors.New("Service discovery not configured")
	}
	return s.Cs.Get(ctx, name)
}

func (s *impl) HTTP() (core.HTTP, error) {
	if s.Mx == nil {
		return nil, errors.New("HTTP service not configured")
	}
	return s, nil
}

func (s *impl) Mux() *http.ServeMux {
	return s.Mx
}

func (s *impl) Locker() (dLocker.DLocker, error) {
	if s.Lk == nil {
		return nil, errors.New("Locker not configured")
	}
	return s.Lk, nil
}

func (s *impl) Events() (events.Events, error) {
	if s.Ev == nil {
		return nil, errors.New("Events not configured")
	}
	return s.Ev, nil
}

func (s *impl) SharedStorage(ns string, cb sharedStorage.Callback) (store.Store, error) {
	if s.ShSt == nil {
		return nil, errors.New("shared storage provider not configured")
	}
	return s.ShSt.SharedStorage(ns, cb)
}
