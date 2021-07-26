package p2pgrpc

import (
	"context"
	"net"

	"github.com/libp2p/go-libp2p-core/host"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-gostream"
	"google.golang.org/grpc"
)

const Protocol protocol.ID = "/grpc/1.0.0"

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
	return grpc.WithContextDialer(func(ctx context.Context, peerIdStr string) (net.Conn, error) {
		pid, err := peer.Decode(peerIdStr)
		if err != nil {
			return nil, err
		}
		return gostream.Dial(ctx, p, pid, Protocol)
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
