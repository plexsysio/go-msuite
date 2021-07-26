package msuite

import (
	"encoding/base64"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/mitchellh/go-homedir"
	"github.com/plexsysio/go-msuite/core"
	"github.com/plexsysio/go-msuite/modules/config"
	"github.com/plexsysio/go-msuite/modules/config/json"
	"github.com/plexsysio/go-msuite/modules/node"
	"path/filepath"
)

type Option func(c config.Config)

func WithGRPC() Option {
	return func(c config.Config) {
		c.Set("UseGRPC", true)
	}
}

func WithGRPCTCPListener(port int) Option {
	return func(c config.Config) {
		c.Set("UseTCP", true)
		c.Set("TCPPort", port)
	}
}

func WithJWT(secret string) Option {
	return func(c config.Config) {
		c.Set("UseJWT", true)
		c.Set("JWTSecret", secret)
	}
}

func WithTracing(name, host string) Option {
	return func(c config.Config) {
		c.Set("UseTracing", true)
		c.Set("TracingName", name)
		c.Set("TracingHost", host)
	}
}

func WithHTTP(port int) Option {
	return func(c config.Config) {
		c.Set("UseHTTP", true)
		c.Set("HTTPPort", port)
	}
}

func WithLocker(lkr string, cfg map[string]string) Option {
	return func(c config.Config) {
		c.Set("UseLocker", true)
		c.Set("Locker", lkr)
		for k, v := range cfg {
			c.Set(k, v)
		}
	}
}

func WithP2PPrivateKey(key crypto.PrivKey) Option {
	return func(c config.Config) {
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
		c.Set("Identity", ident)
	}
}

func WithP2PPort(port int) Option {
	return func(c config.Config) {
		c.Set("UseP2P", true)
		c.Set("SwarmPort", port)
	}
}

func WithRepositoryRoot(path string) Option {
	return func(c config.Config) {
		c.Set("RootPath", path)
	}
}

func WithServiceName(name string) Option {
	return func(c config.Config) {
		c.Set("ServiceName", name)
	}
}

func WithServiceACL(acl map[string]string) Option {
	return func(c config.Config) {
		c.Set("UseACL", true)
		c.Set("ACL", acl)
	}
}

func WithTaskManager(count int) Option {
	return func(c config.Config) {
		c.Set("TMWorkersMin", count)
	}
}

func WithPrometheus(useLatency bool) Option {
	return func(c config.Config) {
		c.Set("UsePrometheus", true)
		if useLatency {
			c.Set("UsePrometheusLatency", true)
		}
	}
}

func WithStaticDiscovery(svcAddrs map[string]string) Option {
	return func(c config.Config) {
		c.Set("UseStaticDiscovery", true)
		c.Set("StaticAddresses", svcAddrs)
	}
}

func defaultOpts(c config.Config) {
	if !c.Exists("ServiceName") {
		c.Set("ServiceName", "msuite")
	}
	if !c.Exists("RootPath") {
		hd, err := homedir.Dir()
		if err != nil {
			panic("Unable to determine home directory")
		}
		c.Set("RootPath", filepath.Join(hd, ".msuite"))
	}
	if c.IsSet("UseP2P") || c.IsSet("UseTCP") || c.IsSet("UseHTTP") {
		var tmCount int
		_ = c.Get("TMWorkersMin", &tmCount)
		tmCount += 1
		if c.IsSet("UseTCP") {
			tmCount += 1
		}
		if c.IsSet("UseP2P") {
			tmCount += 2
		}
		if c.IsSet("UseHTTP") {
			tmCount += 1
		}
		c.Set("TMWorkersMin", tmCount)
	}
}

func New(opts ...Option) (core.Service, error) {
	bCfg := jsonConf.DefaultConfig()
	for _, opt := range opts {
		opt(bCfg)
	}

	defaultOpts(bCfg)

	svc, err := node.New(bCfg)
	if err != nil {
		return nil, err
	}

	return svc, nil
}
