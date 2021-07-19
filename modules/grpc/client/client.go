package grpcclient

import (
	"context"
	"errors"
	"fmt"
	logger "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/plexsysio/go-msuite/modules/config"
	"github.com/plexsysio/go-msuite/modules/grpc/transport/p2pgrpc"
	"github.com/plexsysio/go-msuite/utils"
	"github.com/plexsysio/taskmanager"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"time"
)

var Module = func(c config.Config) fx.Option {
	return fx.Options(
		utils.MaybeProvide(NewP2PClientService, c.IsSet("UseP2P")),
		utils.MaybeProvide(NewStaticClientService, !c.IsSet("UseP2P") && c.IsSet("UseStaticDiscovery")),
	)
}

var log = logger.Logger("grpc/client")

type ClientSvc interface {
	Get(context.Context, string, ...grpc.DialOption) (*grpc.ClientConn, error)
}

func NewP2PClientService(
	svcName string,
	d discovery.Discovery,
	h host.Host,
	tm *taskmanager.TaskManager,
) (ClientSvc, error) {
	csvc := &clientImpl{
		ds:  d,
		h:   h,
		svc: svcName,
	}
	// Start discovery provider
	dp := &discoveryProvider{impl: csvc}
	_, err := tm.Go(dp)
	if err != nil {
		return nil, err
	}
	return csvc, nil
}

type discoveryProvider struct {
	impl *clientImpl
}

func (d *discoveryProvider) Name() string {
	return "DiscoveryProvider"
}

func (d *discoveryProvider) Execute(ctx context.Context) error {
	for {
		log.Infof("Advertising service: %s", d.impl.svc)
		ttl, err := d.impl.ds.Advertise(ctx, d.impl.svc, discovery.TTL(time.Minute*15))
		if err != nil {
			log.Debugf("Error advertising %s: %s", d.impl.svc, err.Error())
			select {
			case <-time.After(time.Minute * 2):
				continue
			case <-ctx.Done():
				return nil
			}
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
	ds  discovery.Discovery
	h   host.Host
	svc string
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
