package grpcclient

import (
	"context"
	"errors"
	"fmt"
	"time"

	logger "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/plexsysio/go-msuite/modules/config"
	"github.com/plexsysio/go-msuite/modules/grpc/p2pgrpc"
	"github.com/plexsysio/taskmanager"
	"google.golang.org/grpc"
)

var log = logger.Logger("grpc/client")

const discoveryTTL = 15 * time.Minute

var ErrNoPeerForSvc = errors.New("failed to find any usable peer for service")

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
		var err error
		for _, svc := range d.services {
			log.Debugf("Advertising service: %s", svc)
			_, err = d.ds.Advertise(ctx, svc, discovery.TTL(discoveryTTL))
			if err != nil {
				err = fmt.Errorf("error advertising %s: %w", svc, err)
				break
			}
		}
		if err != nil {
			log.Errorf("error advertising %v", err)
			select {
			case <-time.After(time.Minute * 2):
				continue
			case <-ctx.Done():
				return nil
			}
		}
		wait := 7 * discoveryTTL / 8
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			log.Info("stopping advertiser")
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

	// FindPeers is called without limit opt, so this cancel is required to release
	// any resources used by it
	cCtx, cCancel := context.WithCancel(ctx)
	defer cCancel()

	p, err := c.ds.FindPeers(cCtx, svc)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case pAddr, more := <-p:
			if !more {
				return nil, ErrNoPeerForSvc
			}
			err = c.h.Connect(ctx, pAddr)
			if err != nil {
				log.Errorf("failed to connect to peer %v err %v", pAddr, err)
				continue
			}
			log.Debugf("connected to peer %v for service %s", pAddr, svc)
			return p2pgrpc.NewP2PDialer(c.h).Dial(ctx, pAddr.ID.String(), opts...)
		}
	}
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
