package p2pgrpc

import (
	"context"
	logger "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/host"
	inet "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
	manet "github.com/multiformats/go-multiaddr-net"
	"go.uber.org/fx"
	"io"
	"net"
)

var log = logger.Logger("transport/p2p")

const Protocol protocol.ID = "/grpc/1.0.0"

func NewP2PListener(lc fx.Lifecycle, h host.Host) (net.Listener, error) {
	p := &p2pListener{
		Host:     h,
		streamCh: make(chan inet.Stream),
	}
	p.listenerCtx, p.listenerCancel = context.WithCancel(context.Background())
	h.SetStreamHandler(Protocol, p.handleStream)
	log.Info("Started listener on P2P Host")
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping listener")
			p.listenerCancel()
			return nil
		},
	})
	return p, nil
}

type p2pListener struct {
	host.Host
	listenerCtx    context.Context
	listenerCancel context.CancelFunc
	streamCh       chan inet.Stream
}

func (p *p2pListener) handleStream(s inet.Stream) {
	select {
	case <-p.listenerCtx.Done():
		return
	case p.streamCh <- s:
	}
}

func (p *p2pListener) Accept() (net.Conn, error) {
	select {
	case <-p.listenerCtx.Done():
		return nil, io.EOF
	case newStream := <-p.streamCh:
		return &p2pConn{Stream: newStream}, nil
	}
}

func (p *p2pListener) Close() error {
	p.listenerCancel()
	return nil
}

func (p *p2pListener) Addr() net.Addr {
	listenAddrs := p.Network().ListenAddresses()
	if len(listenAddrs) > 0 {
		for _, addr := range listenAddrs {
			if na, err := manet.ToNetAddr(addr); err == nil {
				return na
			}
		}
	}
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0}
}

var _ net.Listener = &p2pListener{}
