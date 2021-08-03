package grpcclient

import (
	"context"
	"errors"
	"fmt"
	logger "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/plexsysio/go-msuite/modules/config"
	"github.com/plexsysio/go-msuite/modules/grpc/p2pgrpc"
	"github.com/plexsysio/taskmanager"
	"google.golang.org/grpc"
	"time"
)

var log = logger.Logger("grpc/client")

type ClientSvc interface {
	Get(context.Context, string, ...grpc.DialOption) (*grpc.ClientConn, error)
}

func NewP2PClientService(
	d discovery.Discovery,
	h host.Host,
) (ClientSvc, error) {
	csvc := &clientImpl{
		ds: d,
		h:  h,
	}
	return csvc, nil
}

func NewP2PClientAdvertiser(
	cfg config.Config,
	d discovery.Discovery,
	tm *taskmanager.TaskManager,
) error {
	var services []string
	found := cfg.Get("Services", &services)
	if found {
		// Start discovery provider
		dp := &discoveryProvider{ds: d, services: services}
		_, err := tm.Go(dp)
		if err != nil {
			return err
		}
	}
	return nil
}

type discoveryProvider struct {
	services []string
	ds       discovery.Discovery
}

func (d *discoveryProvider) Name() string {
	return "DiscoveryProvider"
}

func (d *discoveryProvider) Execute(ctx context.Context) error {
	for {
		var (
			startTTL time.Duration
			err      error
		)
		started := time.Now()
		for i, svc := range d.services {
			log.Infof("Advertising service: %s", svc)
			ttl, e := d.ds.Advertise(ctx, svc, discovery.TTL(time.Minute*15))
			if e != nil {
				err = fmt.Errorf("error advertising %s: %w", svc, e)
				break
			}
			// Use TTL of first advertisement for wait in the next part
			if i == 0 {
				startTTL = ttl
			}
		}
		if err != nil {
			log.Debug(err.Error())
			select {
			case <-time.After(time.Minute * 2):
				continue
			case <-ctx.Done():
				return nil
			}
		}
		// Time to wait needs to obey TTL of the first service advertised.
		// If the operation takes time and we wait for all services to advertise, initial
		// services might not get advertised.
		ttl := startTTL - time.Since(started)
		if ttl <= 0 {
			ttl = startTTL
		}
		wait := 7 * ttl / 8
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			log.Info("Stopping advertiser")
			return nil
		}
	}
}

type clientImpl struct {
	ds discovery.Discovery
	h  host.Host
}

func (c *clientImpl) Get(
	ctx context.Context,
	svc string,
	opts ...grpc.DialOption,
) (*grpc.ClientConn, error) {
	p, err := c.ds.FindPeers(ctx, svc, discovery.Limit(1))
	if err != nil {
		return nil, err
	}
	select {
	case <-time.After(time.Second * 10):
		return nil, errors.New("unable to find peer for service " + svc)
	case pAddr, ok := <-p:
		if ok {
			err = c.h.Connect(ctx, pAddr)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to peer %v", pAddr)
			}
			log.Infof("Connected to peer %v for service %s", pAddr, svc)
			return p2pgrpc.NewP2PDialer(c.h).Dial(ctx, pAddr.ID.String(), opts...)
		}
	}
	return nil, errors.New("invalid address received for peer")
}

func NewStaticClientService(c config.Config) ClientSvc {
	svcAddrs := make(map[string]string)
	c.Get("StaticAddresses", &svcAddrs)
	return &staticClientImpl{
		svcAddrs: svcAddrs,
	}
}

type staticClientImpl struct {
	svcAddrs map[string]string
}

func (c *staticClientImpl) Get(
	ctx context.Context,
	svc string,
	opts ...grpc.DialOption,
) (*grpc.ClientConn, error) {
	addr, ok := c.svcAddrs[svc]
	if !ok {
		return nil, errors.New("service address not configured")
	}
	return grpc.DialContext(ctx, addr, opts...)
}
