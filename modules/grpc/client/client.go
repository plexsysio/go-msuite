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
	"sync"
	"time"
)

var Module = fx.Provide(NewClientService)

var log = logger.Logger("grpc/client")

type Client interface {
	Get(context.Context, string, ...grpc.DialOption) (*grpc.ClientConn, error)
}

func NewClientService(
	d discovery.Discovery,
	h host.Host,
	tm *taskmanager.TaskManager,
) *ClientSvc {
	csvc := &ClientSvc{
		d:    d,
		h:    h,
		svcs: []string{},
	}
	// Start discovery provider
	dp := &discoveryProvider{ds: d, impl: csvc}
	tm.GoWork(dp)
	return csvc
}

type ClientSvc struct {
	d    discovery.Discovery
	h    host.Host
	svcs []string
	mtx  sync.Mutex
}

type discoveryProvider struct {
	ds   discovery.Discovery
	impl *ClientSvc
}

func (d *discoveryProvider) Name() string {
	return "DiscoveryProvider"
}

func (d *discoveryProvider) Execute(ctx context.Context) error {
	for {
		ttl := time.Minute * 15
		svcs := d.impl.getSvcs()
		log.Infof("Advertising services No. of Svcs:%d", len(svcs))
		for _, s := range svcs {
			select {
			case <-ctx.Done():
				log.Info("Stopping advertiser")
				return nil
			default:
			}
			_, err := d.ds.Advertise(ctx, s, discovery.TTL(ttl))
			if err != nil {
				log.Debugf("Error advertising %s: %s", s, err.Error())
				continue
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

func (c *ClientSvc) getSvcs() []string {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	cl := make([]string, len(c.svcs))
	copy(cl, c.svcs)
	return cl
}

func (c *ClientSvc) NewClient(
	ctx context.Context,
	name string,
) (Client, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.svcs = append(c.svcs, name)
	return &clientImpl{ds: c.d, h: c.h}, nil
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
