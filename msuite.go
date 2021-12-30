package msuite

import (
	"encoding/base64"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/plexsysio/go-msuite/core"
	"github.com/plexsysio/go-msuite/modules/config"
	jsonConf "github.com/plexsysio/go-msuite/modules/config/json"
	"github.com/plexsysio/go-msuite/modules/node"
)

type BuildCfg struct {
	startupCfg config.Config
}

type Option func(c *BuildCfg)

func WithGRPC(nw string, id interface{}) Option {
	return func(c *BuildCfg) {
		c.startupCfg.Set("UseGRPC", true)
		switch nw {
		case "tcp":
			c.startupCfg.Set("UseTCP", true)
			c.startupCfg.Set("TCPPort", id.(int))
		case "p2p":
			c.startupCfg.Set("UseP2PGRPC", true)
		case "unix":
			c.startupCfg.Set("UseUDS", true)
			c.startupCfg.Set("UDSocket", id.(string))
		}
	}
}

func WithAuth(secret string) Option {
	return func(c *BuildCfg) {
		c.startupCfg.Set("UseAuth", true)
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
		skbytes, err := crypto.MarshalPrivateKey(key)
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

func WithP2P(port int) Option {
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

func WithServices(services ...string) Option {
	return func(c *BuildCfg) {
		c.startupCfg.Set("Services", services)
	}
}

func WithServiceACL(acl map[string]string) Option {
	return func(c *BuildCfg) {
		existingAcls := map[string]string{}
		_ = c.startupCfg.Get("ACL", &existingAcls)
		for k, v := range acl {
			existingAcls[k] = v
		}
		c.startupCfg.Set("ACL", existingAcls)
	}
}

func WithTaskManager(min, max int) Option {
	return func(c *BuildCfg) {
		if max < 20 {
			max += 20
		}
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

func WithDebug() Option {
	return func(c *BuildCfg) {
		c.startupCfg.Set("UseDebug", true)
	}
}

func WithFiles() Option {
	return func(c *BuildCfg) {
		c.startupCfg.Set("UseFiles", true)
	}
}

func WithBootstrapNodes(addrs ...string) Option {
	return func(c *BuildCfg) {
		if len(addrs) > 0 {
			c.startupCfg.Set("BootstrapAddresses", addrs)
		}
	}
}

func defaultOpts(c *BuildCfg) {
	if !c.startupCfg.Exists("Services") {
		c.startupCfg.Set("Services", []string{"msuite"})
	}
}

func New(opts ...Option) (core.Service, error) {
	bCfg := &BuildCfg{
		startupCfg: jsonConf.DefaultConfig(),
	}
	for _, opt := range opts {
		opt(bCfg)
	}

	defaultOpts(bCfg)

	svc, err := node.New(bCfg.startupCfg)
	if err != nil {
		return nil, err
	}

	return svc, nil
}
