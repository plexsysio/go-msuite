package libp2p

import (
	"context"
	"github.com/libp2p/go-libp2p-core/host"
	peer "github.com/libp2p/go-libp2p-peer"
	ps "github.com/libp2p/go-libp2p-peerstore"
	"google.golang.org/grpc"
	"net"
	"time"
)

type P2PGrpcDialer interface {
	Dial(context.Context, string, ...grpc.DialOption) (*grpc.ClientConn, error)
}

func NewP2PDialer(h host.Host) P2PGrpcDialer {
	return &p2pDialer{
		Host: h,
	}
}

type p2pDialer struct {
	host.Host
}

func (p *p2pDialer) getDialer(ctx context.Context) grpc.DialOption {
	return grpc.WithDialer(func(peerIdStr string, timeout time.Duration) (net.Conn, error) {
		subCtx, subCtxCancel := context.WithTimeout(ctx, timeout)
		defer subCtxCancel()

		id, err := peer.IDB58Decode(peerIdStr)
		if err != nil {
			return nil, err
		}
		err = p.Connect(subCtx, ps.PeerInfo{
			ID: id,
		})
		if err != nil {
			return nil, err
		}
		stream, err := p.NewStream(ctx, id, Protocol)
		if err != nil {
			return nil, err
		}
		return &p2pConn{Stream: stream}, nil
	})
}

func (p *p2pDialer) Dial(
	ctx context.Context,
	peerId string,
	dialOpts ...grpc.DialOption,
) (*grpc.ClientConn, error) {
	newOpts := append([]grpc.DialOption{p.getDialer(ctx)}, dialOpts...)
	return grpc.DialContext(ctx, peerId, newOpts...)
}
