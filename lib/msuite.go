package msuite

import (
	"context"
	"github.com/StreamSpace/ss-store"
	"github.com/aloknerurkar/dLocker"
	"github.com/aloknerurkar/go-msuite/modules/auth"
	"github.com/aloknerurkar/go-msuite/modules/cdn"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/aloknerurkar/go-msuite/modules/config/json"
	"github.com/aloknerurkar/go-msuite/modules/events"
	"github.com/aloknerurkar/go-msuite/modules/grpc"
	"github.com/aloknerurkar/go-msuite/modules/grpc/client"
	"github.com/aloknerurkar/go-msuite/modules/http"
	"github.com/aloknerurkar/go-msuite/modules/ipfs"
	"github.com/aloknerurkar/go-msuite/modules/locker"
	"github.com/aloknerurkar/go-msuite/modules/repo"
	"github.com/aloknerurkar/go-msuite/modules/repo/fsrepo"
	ipfslite "github.com/hsanjuan/ipfs-lite"
	ds "github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/routing"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/mitchellh/go-homedir"
	"go.uber.org/fx"
	"google.golang.org/grpc"
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

type impl struct {
	*fx.App
	pCtx   context.Context
	cancel context.CancelFunc

	r    repo.Repo
	h    host.Host
	dht  routing.Routing
	p    *ipfslite.Peer
	ps   *pubsub.PubSub
	disc discovery.Discovery
	st   store.Store
	l    dLocker.DLocker
	rsrv *grpc.Server
}

func (s *impl) Repo() repo.Repo {
	return s.r
}

func (s *impl) Host() host.Host {
	return s.h
}

func (s *impl) Routing() routing.Routing {
	return s.dht
}

func (s *impl) Peer() *ipfslite.Peer {
	return s.p
}

func (s *impl) Pubsub() *pubsub.PubSub {
	return s.ps
}

func (s *impl) Discovery() discovery.Discovery {
	return s.disc
}

func (s *impl) Node() Node {
	return s
}

func (s *impl) Storage() store.Store {
	return s.st
}

func (s *impl) Locker() dLocker.DLocker {
	return s.l
}

func (s *impl) GRPCServer() *grpc.Server {
	return s.rsrv
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
		fx.Provide(func() (context.Context, context.CancelFunc) {
			return context.WithCancel(ctx)
		}),
		fx.Provide(func() (repo.Repo, config.Config, ds.Batching) {
			return r, r.Config(), r.Datastore()
		}),
		ipfs.Module,
		locker.Module,
		auth.Module(r.Config()),
		grpcServer.Module(r.Config()),
		http.Module(r.Config()),
		cdn.Module,
		grpcclient.Module,
		events.Module,
		fx.Populate(svc),
	)

	svc.App = app
	return svc, nil
}
