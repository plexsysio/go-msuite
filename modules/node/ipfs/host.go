package ipfs

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/discovery"
	host "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/routing"
	p2pdiscovery "github.com/libp2p/go-libp2p-discovery"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2ptls "github.com/libp2p/go-libp2p-tls"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/plexsysio/go-msuite/modules/config"
	"github.com/plexsysio/go-msuite/modules/diag/status"
	"github.com/plexsysio/taskmanager"
	"go.uber.org/fx"
)

func Identity(conf config.Config) (crypto.PrivKey, error) {
	id := map[string]interface{}{}
	ok := conf.Get("Identity", &id)
	if !ok {
		return nil, errors.New("Identity info missing")
	}
	privKeyStr, ok := id["PrivKey"]
	if !ok {
		return nil, errors.New("Private key missing")
	}
	pkBytes, err := base64.StdEncoding.DecodeString(privKeyStr.(string))
	if err != nil {
		return nil, err
	}
	priv, err := crypto.UnmarshalPrivateKey(pkBytes)
	if err != nil {
		return nil, err
	}
	return priv, nil
}

var Libp2pOptionsExtra = []libp2p.Option{
	libp2p.NATPortMap(),
	libp2p.ConnectionManager(func() *connmgr.BasicConnMgr {
		connMgr, _ := connmgr.NewConnManager(100, 500, connmgr.WithGracePeriod(time.Minute))
		return connMgr
	}()),
	libp2p.EnableAutoRelay(),
	libp2p.EnableNATService(),
	libp2p.Security(libp2ptls.ID, libp2ptls.New),
}

func Libp2p(
	ctx context.Context,
	lc fx.Lifecycle,
	conf config.Config,
	priv crypto.PrivKey,
) (host.Host, routing.Routing, error) {
	var swPort int
	ok := conf.Get("SwarmPort", &swPort)
	if !ok {
		return nil, nil, errors.New("Swarm Port missing")
	}
	tcpAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", swPort))
	if err != nil {
		return nil, nil, errors.New("Invalid swarm port Err:" + err.Error())
	}
	listenAddrs := []multiaddr.Multiaddr{tcpAddr}
	h, dht, err := ipfslite.SetupLibp2p(
		ctx,
		priv,
		nil,
		listenAddrs,
		nil,
		Libp2pOptionsExtra...,
	)
	if err != nil {
		return nil, nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(c context.Context) error {
			h.Close()
			dht.Close()
			return nil
		},
	})
	return h, dht, nil
}

func LocalDialer(
	lc fx.Lifecycle,
) (host.Host, error) {
	h, err := libp2p.New(
		libp2p.DefaultTransports,
		libp2p.NoListenAddrs,
	)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return h.Close()
		},
	})
	log.Debug("created local dialer", h.ID())
	return h, nil
}

func NewNode(
	ctx context.Context,
	h host.Host,
	dht routing.Routing,
	rootDS datastore.Batching,
) (*ipfslite.Peer, error) {
	return ipfslite.New(
		ctx,
		rootDS,
		h,
		dht,
		&ipfslite.Config{
			Offline: false,
		},
	)
}

func Pubsub(ctx context.Context, h host.Host) (*pubsub.PubSub, error) {
	return pubsub.NewGossipSub(ctx, h, pubsub.WithFloodPublish(true))
}

func NewSvcDiscovery(r routing.Routing) discovery.Discovery {
	return p2pdiscovery.NewRoutingDiscovery(r)
}

func NewP2PReporter(h host.Host, st status.Manager) {
	st.AddReporter("P2P Service", &p2pReporter{h: h})
}

type p2pReporter struct {
	h host.Host
}

func (p *p2pReporter) Status() interface{} {
	stat := make(map[string]interface{})

	stat["ID"] = p.h.ID()
	stat["Addrs"] = p.h.Addrs()
	stat["Peers"] = p.h.Network().Peers()

	return stat
}

func parseBootstrapPeers(addrs []string) ([]peer.AddrInfo, error) {
	maddrs := make([]multiaddr.Multiaddr, len(addrs))
	for i, addr := range addrs {
		var err error
		maddrs[i], err = multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
	}
	return peer.AddrInfosFromP2pAddrs(maddrs...)
}

func Bootstrapper(
	lc fx.Lifecycle,
	cfg config.Config,
	tm *taskmanager.TaskManager,
	h host.Host,
) error {
	var addrs []string
	if cfg.Get("BootstrapAddresses", &addrs) {
		peers, err := parseBootstrapPeers(addrs)
		if err != nil {
			return err
		}

		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				sched, err := tm.GoFunc("Bootstrapper", func(c context.Context) error {
					t := time.NewTicker(15 * time.Second)
					for {
						select {
						case <-c.Done():
							return nil
						case <-t.C:
							for _, p := range peers {
								if h.Network().Connectedness(p.ID) != network.Connected {
									h.Peerstore().AddAddrs(p.ID, p.Addrs, peerstore.PermanentAddrTTL)
									if err := h.Connect(ctx, p); err != nil {
										log.Warn("could not connect to bootstrap address", p)
									}
								}
							}
						}
					}
				})
				if err != nil {
					return err
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-sched:
					return nil
				}
			},
		})
	}

	return nil
}
