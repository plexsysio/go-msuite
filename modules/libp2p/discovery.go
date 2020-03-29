package libp2p

import (
	"context"
	host "github.com/libp2p/go-libp2p-host"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	"time"
)

func NewMDNSDiscovery(h host.Host) error {
	ser, err := discovery.NewMdnsService(context.Background(), h,
		time.Minute*5, Rendezvous)
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

func (d *discoveryNotifiee) HandlePeerFound(pi pstore.PeerInfo) {
	log.Infof("Peer discovery %s", pi.ID.Pretty())
	d.Peerstore().AddAddrs(pi.ID, pi.Addrs, pstore.PermanentAddrTTL)
	err := d.Connect(context.Background(), pi)
	if err != nil {
		log.Errorf("Error connecting to discovered node Err:%s", err.Error())
		return
	}
	log.Infof("Successfully connected to peer %s", pi.ID.Pretty())
}
