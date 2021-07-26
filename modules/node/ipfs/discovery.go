package ipfs

import (
	"context"
	"time"

	logger "github.com/ipfs/go-log/v2"
	host "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pstore "github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery"
)

var log = logger.Logger("mdnsdiscovery")

const Rendezvous string = "/msuite/node"

func NewMDNSDiscovery(ctx context.Context, h host.Host) error {
	ser, err := discovery.NewMdnsService(ctx, h, time.Minute*5, Rendezvous)
	if err != nil {
		log.Errorf("Failed registering MDNS service Err:%s", err.Error())
		return err
	}
	ser.RegisterNotifee(&discoveryNotifiee{Host: h})
	return nil
}

type discoveryNotifiee struct {
	host.Host
}

func (d *discoveryNotifiee) HandlePeerFound(pi peer.AddrInfo) {
	log.Infof("Peer discovery %s", pi.ID.Pretty())
	d.Peerstore().AddAddrs(pi.ID, pi.Addrs, pstore.PermanentAddrTTL)
	err := d.Connect(context.Background(), pi)
	if err != nil {
		log.Errorf("Error connecting to discovered node Err:%s", err.Error())
		return
	}
	log.Infof("Successfully connected to peer %s", pi.ID.Pretty())
}
