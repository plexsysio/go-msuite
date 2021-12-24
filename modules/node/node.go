package node

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	ipfslite "github.com/hsanjuan/ipfs-lite"
	ds "github.com/ipfs/go-datastore"
	logger "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/routing"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/plexsysio/dLocker"
	store "github.com/plexsysio/gkvstore"
	"github.com/plexsysio/go-msuite/core"
	"github.com/plexsysio/go-msuite/modules/auth"
	"github.com/plexsysio/go-msuite/modules/config"
	"github.com/plexsysio/go-msuite/modules/diag/metrics"
	"github.com/plexsysio/go-msuite/modules/diag/status"
	"github.com/plexsysio/go-msuite/modules/events"
	grpcclient "github.com/plexsysio/go-msuite/modules/grpc/client"
	grpcsvc "github.com/plexsysio/go-msuite/modules/node/grpc"
	mhttp "github.com/plexsysio/go-msuite/modules/node/http"
	"github.com/plexsysio/go-msuite/modules/node/ipfs"
	"github.com/plexsysio/go-msuite/modules/node/locker"
	"github.com/plexsysio/go-msuite/modules/repo"
	"github.com/plexsysio/go-msuite/modules/repo/fsrepo"
	"github.com/plexsysio/go-msuite/modules/repo/inmem"
	"github.com/plexsysio/go-msuite/modules/sharedStorage"
	"github.com/plexsysio/go-msuite/utils"
	"github.com/plexsysio/taskmanager"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type FxLog struct{}

var log = logger.Logger("node")

func (*FxLog) Printf(msg string, args ...interface{}) {
	log.Infof(msg, args...)
}

func New(bCfg config.Config) (core.Service, error) {
	var (
		r   repo.Repo
		err error
	)
	if bCfg.Get("RootPath", new(string)) {
		r, err = fsrepo.CreateOrOpen(bCfg)
		if err != nil {
			return nil, err
		}
	} else {
		r, err = inmem.CreateOrOpen(bCfg)
		if err != nil {
			return nil, err
		}
	}

	// if the repo was initialized, the passed config has been saved in it and hence
	// it should return the same values. if the repo was already initialized, this
	// would get the saved config. This allows for node to be started with just the
	// root path
	bCfg = r.Config()

	svc := &impl{}
	dp := deps{}

	app := fx.New(
		fx.Logger(&FxLog{}),
		fx.Provide(func() (context.Context, context.CancelFunc) {
			return context.WithCancel(context.Background())
		}),
		fx.Provide(func(lc fx.Lifecycle) (repo.Repo, config.Config, ds.Batching) {
			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					log.Debugf("stopping repo")
					defer log.Debugf("stopped repo")
					return r.Close()
				},
			})
			return r, r.Config(), r.Datastore()
		}),
		fx.Provide(NewTaskManager),
		fx.Provide(status.New),
		utils.MaybeProvide(metrics.New, bCfg.IsSet("UsePrometheus")),
		utils.MaybeProvide(metrics.NewTracer, bCfg.IsSet("UseTracing")),
		utils.MaybeOption(locker.Module, bCfg.IsSet("UseLocker")),
		utils.MaybeOption(auth.Module, bCfg.IsSet("UseAuth")),
		utils.MaybeOption(ipfs.P2PModule, bCfg.IsSet("UseP2P")),
		utils.MaybeOption(ipfs.FilesModule, bCfg.IsSet("UseP2P") && bCfg.IsSet("UseFiles")),
		utils.MaybeOption(grpcsvc.Module(r.Config()), bCfg.IsSet("UseGRPC")),
		utils.MaybeOption(mhttp.Module(r.Config()), bCfg.IsSet("UseHTTP")),
		utils.MaybeOption(fx.Provide(events.NewEventsSvc), bCfg.IsSet("UseP2P")),
		utils.MaybeOption(fx.Provide(sharedStorage.NewSharedStoreProvider), bCfg.IsSet("UseP2P")),
		utils.MaybeInvoke(status.RegisterHTTP, bCfg.IsSet("UseHTTP")),
		fx.Invoke(func(lc fx.Lifecycle, cancel context.CancelFunc) {
			lc.Append(fx.Hook{
				OnStop: func(c context.Context) error {
					cancel()
					return nil
				},
			})
		}),
		fx.Invoke(func(c config.Config, tm *taskmanager.TaskManager, st status.Manager) {
			st.AddReporter("Repository", r)
			st.AddReporter("TaskManager", &tmReporter{tm})
			st.AddReporter("Services", &svcsReporter{c})
		}),
		fx.Populate(&dp),
	)

	svc.App = app
	svc.dp = dp
	return svc, nil
}

func NewTaskManager(lc fx.Lifecycle, cfg config.Config) (*taskmanager.TaskManager, error) {
	tmCfg := map[string]int{}
	found := cfg.Get("TMWorkers", &tmCfg)
	if !found {
		tmCfg["Min"] = 0
		tmCfg["Max"] = 20
	}
	if tmCfg["Max"] <= 0 {
		return nil, errors.New("invalid config for taskmanager workers")
	}
	tm := taskmanager.New(tmCfg["Min"], tmCfg["Max"], time.Second*15)
	lc.Append(fx.Hook{
		OnStop: func(c context.Context) error {
			log.Debugf("stopping taskmanager")
			defer log.Debugf("stopped taskmanager")
			tm.Stop()
			return nil
		},
	})
	return tm, nil
}

type tmReporter struct {
	tm *taskmanager.TaskManager
}

func (t *tmReporter) Status() interface{} { return t.tm.Status() }

type svcsReporter struct {
	c config.Config
}

func (s *svcsReporter) Status() interface{} {
	var svcs []string
	found := s.c.Get("Services", &svcs)
	if !found {
		return "no services configured"
	}
	return svcs
}

type deps struct {
	fx.In

	Ctx    context.Context
	Cancel context.CancelFunc
	R      repo.Repo
	Am     auth.ACL                 `optional:"true"`
	Tm     *taskmanager.TaskManager `optional:"true"`
	Lk     dLocker.DLocker          `optional:"true"`
	Rsrv   *grpc.Server             `optional:"true"`
	Mx     *http.ServeMux           `optional:"true"`
	Gmx    *runtime.ServeMux        `optional:"true"`
	H      host.Host                `optional:"true"`
	Dht    routing.Routing          `optional:"true"`
	P      *ipfslite.Peer           `optional:"true"`
	Ps     *pubsub.PubSub           `optional:"true"`
	Disc   discovery.Discovery      `optional:"true"`
	St     store.Store              `optional:"true"`
	Jm     auth.JWTManager          `optional:"true"`
	Ev     events.Events            `optional:"true"`
	Cs     grpcclient.ClientSvc     `optional:"true"`
	ShSt   sharedStorage.Provider   `optional:"true"`
}

type impl struct {
	*fx.App
	dp deps
}

func (s *impl) Repo() repo.Repo {
	return s.dp.R
}

func (s *impl) TM() *taskmanager.TaskManager {
	return s.dp.Tm
}

func (s *impl) P2P() (core.P2P, error) {
	if s.dp.H == nil || s.dp.Dht == nil || s.dp.Disc == nil || s.dp.Ps == nil {
		return nil, errors.New("P2P not configured")
	}
	return s, nil
}

func (s *impl) Host() host.Host {
	return s.dp.H
}

func (s *impl) Routing() routing.Routing {
	return s.dp.Dht
}

func (s *impl) Discovery() discovery.Discovery {
	return s.dp.Disc
}

func (s *impl) Pubsub() *pubsub.PubSub {
	return s.dp.Ps
}

// Files API
func (s *impl) Files() (*ipfslite.Peer, error) {
	if s.dp.P == nil {
		return nil, errors.New("Files service not configured")
	}
	return s.dp.P, nil
}

// Auth API
func (s *impl) Auth() (core.Auth, error) {
	if s.dp.Jm == nil || s.dp.Am == nil {
		return nil, errors.New("Auth not configured")
	}
	return s, nil
}

func (s *impl) JWT() auth.JWTManager {
	return s.dp.Jm
}

func (s *impl) ACL() auth.ACL {
	return s.dp.Am
}

func (s *impl) GRPC() (core.GRPC, error) {
	if s.dp.Rsrv == nil {
		return nil, errors.New("GRPC service not configured")
	}
	return s, nil
}

func (s *impl) Server() *grpc.Server {
	return s.dp.Rsrv
}

func (s *impl) Client(ctx context.Context, name string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if s.dp.Cs == nil {
		return nil, errors.New("Service discovery not configured")
	}
	return s.dp.Cs.Get(ctx, name, opts...)
}

func (s *impl) HTTP() (core.HTTP, error) {
	if s.dp.Mx == nil {
		return nil, errors.New("HTTP service not configured")
	}
	return s, nil
}

func (s *impl) Mux() *http.ServeMux {
	return s.dp.Mx
}

func (s *impl) Gateway() *runtime.ServeMux {
	return s.dp.Gmx
}

func (s *impl) Locker() (dLocker.DLocker, error) {
	if s.dp.Lk == nil {
		return nil, errors.New("Locker not configured")
	}
	return s.dp.Lk, nil
}

func (s *impl) Events() (events.Events, error) {
	if s.dp.Ev == nil {
		return nil, errors.New("Events not configured")
	}
	return s.dp.Ev, nil
}

func (s *impl) SharedStorage(ns string, cb sharedStorage.Callback) (store.Store, error) {
	if s.dp.ShSt == nil {
		return nil, errors.New("shared storage provider not configured")
	}
	return s.dp.ShSt.SharedStorage(ns, cb)
}
