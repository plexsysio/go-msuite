package msuite

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/StreamSpace/ss-store"
	"github.com/StreamSpace/ss-taskmanager"
	"github.com/aloknerurkar/dLocker"
	"github.com/aloknerurkar/go-msuite/modules/auth"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/aloknerurkar/go-msuite/modules/config/json"
	"github.com/aloknerurkar/go-msuite/modules/diag/status"
	"github.com/aloknerurkar/go-msuite/modules/events"
	"github.com/aloknerurkar/go-msuite/modules/grpc"
	"github.com/aloknerurkar/go-msuite/modules/grpc/client"
	mhttp "github.com/aloknerurkar/go-msuite/modules/http"
	"github.com/aloknerurkar/go-msuite/modules/ipfs"
	"github.com/aloknerurkar/go-msuite/modules/locker"
	"github.com/aloknerurkar/go-msuite/modules/repo"
	"github.com/aloknerurkar/go-msuite/modules/repo/fsrepo"
	"github.com/aloknerurkar/go-msuite/utils"
	ipfslite "github.com/hsanjuan/ipfs-lite"
	ds "github.com/ipfs/go-datastore"
	logger "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
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

type BuildCfg struct {
	cfg     config.Config
	root    string
	svcName string
	tmCount int
}

type Option func(c *BuildCfg)

func WithGRPCTCPListener(port int) Option {
	return func(c *BuildCfg) {
		c.cfg.Set("UseTCP", true)
		c.cfg.Set("TCPPort", port)
	}
}

func WithJWT(secret string) Option {
	return func(c *BuildCfg) {
		c.cfg.Set("UseJWT", true)
		c.cfg.Set("JWTSecret", secret)
	}
}

func WithTracing(name, host string) Option {
	return func(c *BuildCfg) {
		c.cfg.Set("UseTracing", true)
		c.cfg.Set("TracingName", name)
		c.cfg.Set("TracingHost", host)
	}
}

func WithHTTP(port int) Option {
	return func(c *BuildCfg) {
		c.cfg.Set("UseHTTP", true)
		c.cfg.Set("HTTPPort", port)
	}
}

func WithLocker(lkr string, cfg map[string]string) Option {
	return func(c *BuildCfg) {
		c.cfg.Set("UseLocker", true)
		c.cfg.Set("Locker", lkr)
		for k, v := range cfg {
			c.cfg.Set(k, v)
		}
	}
}

func WithP2PPrivateKey(key crypto.PrivKey) Option {
	return func(c *BuildCfg) {
		skbytes, err := key.Bytes()
		if err != nil {
			return
		}
		ident := map[string]interface{}{}
		ident["PrivKey"] = base64.StdEncoding.EncodeToString(skbytes)

		id, err := peer.IDFromPublicKey(key.GetPublic())
		if err != nil {
			return
		}
		ident["ID"] = id.Pretty()
		c.cfg.Set("Identity", ident)
	}
}

func WithP2PPort(port int) Option {
	return func(c *BuildCfg) {
		c.cfg.Set("UseP2P", true)
		c.cfg.Set("SwarmPort", port)
	}
}

func WithRepositoryRoot(path string) Option {
	return func(c *BuildCfg) {
		c.root = path
	}
}

func WithServiceName(name string) Option {
	return func(c *BuildCfg) {
		c.svcName = name
	}
}

func WithServiceACL(acl map[string]string) Option {
	return func(c *BuildCfg) {
		c.cfg.Set("UseACL", true)
		c.cfg.Set("ACL", acl)
	}
}

func WithTaskManager(count int) Option {
	return func(c *BuildCfg) {
		c.tmCount = count
	}
}

func defaultOpts(c *BuildCfg) {
	if len(c.svcName) == 0 {
		c.svcName = "msuite"
	}
	if len(c.root) == 0 {
		hd, err := homedir.Dir()
		if err != nil {
			panic("Unable to determine home directory")
		}
		c.root = filepath.Join(hd, ".msuite")
	}
	if c.cfg.IsSet("UseP2P") || c.cfg.IsSet("UseTCP") {
		c.tmCount += 1
		if c.cfg.IsSet("UseTCP") {
			c.tmCount += 1
		}
		if c.cfg.IsSet("UseP2P") {
			c.tmCount += 2
		}
	}
}

func New(opts ...Option) (Service, error) {
	bCfg := &BuildCfg{
		cfg: jsonConf.DefaultConfig(),
	}
	for _, opt := range opts {
		opt(bCfg)
	}
	defaultOpts(bCfg)
	r, err := fsrepo.CreateOrOpen(bCfg.root, bCfg.cfg)
	if err != nil {
		return nil, err
	}
	svc := &impl{}

	app := fx.New(
		fx.Logger(&FxLog{}),
		fx.Provide(func() (context.Context, context.CancelFunc) {
			return context.WithCancel(context.Background())
		}),
		fx.Provide(func() (repo.Repo, config.Config, ds.Batching) {
			return r, r.Config(), r.Datastore()
		}),
		utils.MaybeProvide(func(ctx context.Context) *taskmanager.TaskManager {
			return taskmanager.NewTaskManager(ctx, int32(bCfg.tmCount))
		}, bCfg.tmCount > 0),
		fx.Provide(func() string {
			return bCfg.svcName
		}),
		status.Module,
		utils.MaybeOption(locker.Module, bCfg.cfg.IsSet("UseLocker")),
		auth.Module(r.Config()),
		utils.MaybeOption(ipfs.Module, bCfg.cfg.IsSet("UseP2P")),
		utils.MaybeOption(grpcServer.Module(r.Config()),
			bCfg.cfg.IsSet("UseTCP") || bCfg.cfg.IsSet("UseP2P")),
		mhttp.Module(r.Config()),
		utils.MaybeOption(grpcclient.Module, bCfg.cfg.IsSet("UseP2P")),
		utils.MaybeOption(events.Module, bCfg.cfg.IsSet("UseP2P")),
		fx.Invoke(func(lc fx.Lifecycle, cancel context.CancelFunc) {
			lc.Append(fx.Hook{
				OnStop: func(c context.Context) error {
					cancel()
					r.Close()
					return nil
				},
			})
		}),
		utils.MaybeInvoke(func(lc fx.Lifecycle, tm *taskmanager.TaskManager) {
			lc.Append(fx.Hook{
				OnStop: func(c context.Context) error {
					tm.Stop()
					return nil
				},
			})
		}, bCfg.tmCount > 0),
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

func (s *impl) Node() (Node, error) {
	if s.H == nil {
		return nil, errors.New("Node not configured")
	}
	return s, nil
}

func (s *impl) Storage() store.Store {
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

func (s *impl) GRPC() (GRPC, error) {
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

func (s *impl) HTTP() (HTTP, error) {
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
