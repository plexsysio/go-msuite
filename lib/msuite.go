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
	"os"
	"path/filepath"
)

type Service interface {
	Start(context.Context) error
	Stop(context.Context) error
	Done() <-chan os.Signal

	Node() Node
	Storage() store.Store
	Locker() dLocker.DLocker
	GRPCServer() *grpc.Server
}

type Node interface {
	Repo() repo.Repo
	Host() host.Host
	Routing() routing.Routing
	Peer() *ipfslite.Peer
	Pubsub() *pubsub.PubSub
	Discovery() discovery.Discovery
}

type FxLog struct{}

var log = logger.Logger("Boot")

func (f *FxLog) Printf(msg string, args ...interface{}) {
	log.Infof(msg, args...)
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
	Lk   dLocker.DLocker
	Rsrv *grpc.Server
	Tm   *taskmanager.TaskManager
	Ev   events.Events
	Cs   *grpcclient.ClientSvc
	Mx   *http.ServeMux
}

func (s *impl) Repo() repo.Repo {
	return s.R
}

func (s *impl) Host() host.Host {
	return s.H
}

func (s *impl) Routing() routing.Routing {
	return s.Dht
}

func (s *impl) Peer() *ipfslite.Peer {
	return s.P
}

func (s *impl) Pubsub() *pubsub.PubSub {
	return s.Ps
}

func (s *impl) Discovery() discovery.Discovery {
	return s.Disc
}

func (s *impl) Node() Node {
	return s
}

func (s *impl) Storage() store.Store {
	return s.St
}

func (s *impl) Locker() dLocker.DLocker {
	return s.Lk
}

func (s *impl) GRPCServer() *grpc.Server {
	return s.Rsrv
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
		fx.Populate(svc),
	)

	svc.App = app
	return svc, nil
}
