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
	"github.com/libp2p/go-libp2p-core/routing"
	p2pdiscovery "github.com/libp2p/go-libp2p-discovery"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2ptls "github.com/libp2p/go-libp2p-tls"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/plexsysio/go-msuite/modules/config"
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
	libp2p.ConnectionManager(connmgr.NewConnManager(10, 50, time.Minute)),
	libp2p.EnableAutoRelay(),
	libp2p.EnableNATService(),
	libp2p.Security(libp2ptls.ID, libp2ptls.New),
	libp2p.DefaultTransports,
}

func Libp2p(
	ctx context.Context,
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
	return ipfslite.SetupLibp2p(
		ctx,
		priv,
		nil,
		listenAddrs,
		nil,
		Libp2pOptionsExtra...,
	)
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
	return pubsub.NewGossipSub(ctx, h)
}

func NewSvcDiscovery(r routing.Routing) discovery.Discovery {
	return p2pdiscovery.NewRoutingDiscovery(r)
}
