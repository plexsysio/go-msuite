package msuite

import (
	"context"
	"encoding/base64"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/mitchellh/go-homedir"
	"github.com/plexsysio/go-msuite/core"
	"github.com/plexsysio/go-msuite/modules/config"
	"github.com/plexsysio/go-msuite/modules/config/json"
	"github.com/plexsysio/go-msuite/modules/node"
	"path/filepath"
	"time"
)

type BuildCfg struct {
	startupCfg config.Config
	services   map[string]func(core.Service) error
}

type Option func(c *BuildCfg)

func WithGRPC() Option {
	return func(c *BuildCfg) {
		c.startupCfg.Set("UseGRPC", true)
	}
}

func WithGRPCTCPListener(port int) Option {
	return func(c *BuildCfg) {
		c.startupCfg.Set("UseTCP", true)
		c.startupCfg.Set("TCPPort", port)
	}
}

func WithJWT(secret string) Option {
	return func(c *BuildCfg) {
		c.startupCfg.Set("UseJWT", true)
		c.startupCfg.Set("JWTSecret", secret)
	}
}

func WithTracing(name, host string) Option {
	return func(c *BuildCfg) {
		c.startupCfg.Set("UseTracing", true)
		c.startupCfg.Set("TracingName", name)
		c.startupCfg.Set("TracingHost", host)
	}
}

func WithHTTP(port int) Option {
	return func(c *BuildCfg) {
		c.startupCfg.Set("UseHTTP", true)
		c.startupCfg.Set("HTTPPort", port)
	}
}

func WithLocker(lkr string, cfg map[string]string) Option {
	return func(c *BuildCfg) {
		c.startupCfg.Set("UseLocker", true)
		c.startupCfg.Set("Locker", lkr)
		for k, v := range cfg {
			c.startupCfg.Set(k, v)
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
		c.startupCfg.Set("Identity", ident)
	}
}

func WithP2PPort(port int) Option {
	return func(c *BuildCfg) {
		c.startupCfg.Set("UseP2P", true)
		c.startupCfg.Set("SwarmPort", port)
	}
}

func WithRepositoryRoot(path string) Option {
	return func(c *BuildCfg) {
		c.startupCfg.Set("RootPath", path)
	}
}

func WithServiceName(name string) Option {
	return func(c *BuildCfg) {
		c.startupCfg.Set("Services", []string{name})
	}
}

func WithServiceACL(acl map[string]string) Option {
	return func(c *BuildCfg) {
		c.startupCfg.Set("UseACL", true)
		c.startupCfg.Set("ACL", acl)
	}
}

func WithTaskManager(min, max int) Option {
	return func(c *BuildCfg) {
		c.startupCfg.Set("TMWorkers", map[string]int{
			"Min": min,
			"Max": max,
		})
	}
}

func WithPrometheus(useLatency bool) Option {
	return func(c *BuildCfg) {
		c.startupCfg.Set("UsePrometheus", true)
		if useLatency {
			c.startupCfg.Set("UsePrometheusLatency", true)
		}
	}
}

func WithStaticDiscovery(svcAddrs map[string]string) Option {
	return func(c *BuildCfg) {
		c.startupCfg.Set("UseStaticDiscovery", true)
		c.startupCfg.Set("StaticAddresses", svcAddrs)
	}
}

func WithService(name string, initFn func(core.Service) error) Option {
	return func(c *BuildCfg) {
		c.services[name] = initFn
	}
}

func defaultOpts(c *BuildCfg) {
	if !c.startupCfg.Exists("Services") {
		c.startupCfg.Set("Services", []string{"msuite"})
	}

	var services []string
	_ = c.startupCfg.Get("Services", &services)
	for k, _ := range c.services {
		services = append(services, k)
	}
	c.startupCfg.Set("Services", services)

	if !c.startupCfg.Exists("RootPath") {
		hd, err := homedir.Dir()
		if err != nil {
			panic("Unable to determine home directory")
		}
		c.startupCfg.Set("RootPath", filepath.Join(hd, ".msuite"))
	}
	if c.startupCfg.IsSet("UseP2P") || c.startupCfg.IsSet("UseTCP") || c.startupCfg.IsSet("UseHTTP") {
		tmCfg := map[string]int{}
		found := c.startupCfg.Get("TMWorkers", &tmCfg)
		if !found {
			tmCfg = map[string]int{
				"Min": 0,
				"Max": 0,
			}
		}
		tmCfg["Min"] += 1
		if c.startupCfg.IsSet("UseTCP") {
			tmCfg["Min"] += 1
		}
		if c.startupCfg.IsSet("UseP2P") {
			tmCfg["Min"] += 2
		}
		if c.startupCfg.IsSet("UseHTTP") {
			tmCfg["Min"] += 1
		}
		if tmCfg["Max"] < tmCfg["Min"] {
			tmCfg["Max"] = 2 * tmCfg["Min"]
		}
		c.startupCfg.Set("TMWorkers", tmCfg)
	}
}

func New(opts ...Option) (core.Service, error) {
	bCfg := &BuildCfg{
		startupCfg: jsonConf.DefaultConfig(),
		services:   make(map[string]func(core.Service) error),
	}
	for _, opt := range opts {
		opt(bCfg)
	}

	defaultOpts(bCfg)

	svc, err := node.New(bCfg.startupCfg)
	if err != nil {
		return nil, err
	}

	for _, initFn := range bCfg.services {
		err = initFn(svc)
		if err != nil {
			ctxd, _ := context.WithTimeout(context.Background(), time.Second*5)
			_ = svc.Stop(ctxd)
			return nil, err
		}
	}

	return svc, nil
}
