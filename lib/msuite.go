package msuite

import (
	"context"
	"github.com/StreamSpace/ss-store"
	"github.com/StreamSpace/ss-taskmanager"
	"github.com/aloknerurkar/dLocker"
	"github.com/aloknerurkar/go-msuite/modules/auth"
	"github.com/aloknerurkar/go-msuite/modules/cdn"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/aloknerurkar/go-msuite/modules/config/json"
	"github.com/aloknerurkar/go-msuite/modules/events"
	"github.com/aloknerurkar/go-msuite/modules/grpc"
	"github.com/aloknerurkar/go-msuite/modules/grpc/client"
	mhttp "github.com/aloknerurkar/go-msuite/modules/http"
	"github.com/aloknerurkar/go-msuite/modules/ipfs"
	"github.com/aloknerurkar/go-msuite/modules/locker"
	"github.com/aloknerurkar/go-msuite/modules/repo"
	"github.com/aloknerurkar/go-msuite/modules/repo/fsrepo"
	ipfslite "github.com/hsanjuan/ipfs-lite"
	ds "github.com/ipfs/go-datastore"
	logger "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/routing"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/mitchellh/go-homedir"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"net/http"
	"path/filepath"
)

type FxLog struct{}

var log = logger.Logger("Boot")

func (f *FxLog) Printf(msg string, args ...interface{}) {
	log.Infof(msg, args...)
}

func New(ctx context.Context) (Service, error) {
	hd, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	rootPath := filepath.Join(hd, ".msuite")
	r, err := fsrepo.CreateOrOpen(rootPath, jsonConf.DefaultConfig())
	if err != nil {
		return nil, err
	}
	svc := &impl{}

	app := fx.New(
		fx.Logger(&FxLog{}),
		fx.Provide(func() (context.Context, context.CancelFunc) {
			return context.WithCancel(ctx)
		}),
		fx.Provide(func() (repo.Repo, config.Config, ds.Batching) {
			return r, r.Config(), r.Datastore()
		}),
		fx.Provide(func(ctx context.Context) *taskmanager.TaskManager {
			return taskmanager.NewTaskManager(ctx, 10)
		}),
		ipfs.Module,
		locker.Module,
		auth.Module(r.Config()),
		grpcServer.Module(r.Config()),
		mhttp.Module(r.Config()),
		cdn.Module,
		grpcclient.Module,
		events.Module,
		fx.Invoke(func(lc fx.Lifecycle, cancel context.CancelFunc) {
			lc.Append(fx.Hook{
				OnStop: func(c context.Context) error {
					cancel()
					r.Close()
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
	H    host.Host
	Dht  routing.Routing
	P    *ipfslite.Peer
	Ps   *pubsub.PubSub
	Disc discovery.Discovery
	St   store.Store
	Jm   auth.JWTManager
	Am   auth.ACL
	Lk   dLocker.DLocker
	Rsrv *grpc.Server
	Tm   *taskmanager.TaskManager
	Ev   events.Events
	Cs   *grpcclient.ClientSvc
	Mx   *http.ServeMux
}

// Node API
func (s *impl) Node() Node {
	return s
}

func (s *impl) Repo() repo.Repo {
	return s.R
}

// Storage API
func (s *impl) Storage() Storage {
	return s
}

func (s *impl) Local() store.Store {
	return s.R.Store()
}

func (s *impl) Shared() store.Store {
	return s.St
}

// P2P API
func (s *impl) P2P() P2P {
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
func (s *impl) Auth() Auth {
	return s
}

func (s *impl) JWT() auth.JWTManager {
	return s.Jm
}

func (s *impl) ACL() auth.ACL {
	return s.Am
}

func (s *impl) GRPC() GRPC {
	return s
}

func (s *impl) Server() *grpc.Server {
	return s.Rsrv
}

func (s *impl) Client(ctx context.Context, name string) (grpcclient.Client, error) {
	return s.Cs.NewClient(ctx, name)
}

func (s *impl) HTTP() HTTP {
	return s
}

func (s *impl) Mux() *http.ServeMux {
	return s.Mx
}

func (s *impl) Locker() dLocker.DLocker {
	return s.Lk
}

func (s *impl) Events() events.Events {
	return s.Ev
}
