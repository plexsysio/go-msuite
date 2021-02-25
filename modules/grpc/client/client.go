package grpcclient

import (
	"context"
	"errors"
	"fmt"
	"github.com/StreamSpace/ss-taskmanager"
	"github.com/aloknerurkar/go-msuite/modules/grpc/transport/p2pgrpc"
	logger "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/host"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"time"
)

var Module = fx.Provide(NewClientService)

var log = logger.Logger("grpc/client")

type ClientSvc interface {
	Get(context.Context, string, ...grpc.DialOption) (*grpc.ClientConn, error)
}

func NewClientService(
	svcName string,
	d discovery.Discovery,
	h host.Host,
	tm *taskmanager.TaskManager,
) ClientSvc {
	csvc := &clientImpl{
		ds:  d,
		h:   h,
		svc: svcName,
	}
	// Start discovery provider
	dp := &discoveryProvider{impl: csvc}
	tm.GoWork(dp)
	return csvc
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
		return nil, errors.New("Unable to find peer for service " + svc)
	case pAddr, ok := <-p:
		if ok {
			err = c.h.Connect(ctx, pAddr)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Failed to connect to peer %v", pAddr))
			}
			log.Infof("Connected to peer %v for service %s", pAddr, svc)
			return p2pgrpc.NewP2PDialer(c.h).Dial(ctx, pAddr.ID.String(), opts...)
		}
	}
	return nil, errors.New("Invalid address received for peer")
}
