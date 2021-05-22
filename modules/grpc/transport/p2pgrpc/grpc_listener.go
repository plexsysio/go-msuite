package p2pgrpc

import (
	"context"

	logger "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-gostream"
	"github.com/plexsysio/go-msuite/modules/grpc/transport/mux"
)

var log = logger.Logger("transport/p2p")

const Protocol protocol.ID = "/grpc/1.0.0"

func NewP2PListener(
	ctx context.Context,
	h host.Host,
) (grpcmux.MuxListenerOut, error) {
	p, err := gostream.Listen(h, Protocol)
	if err != nil {
		return grpcmux.MuxListenerOut{}, err
	}
	log.Info("Started listener on P2P Host")
	return grpcmux.MuxListenerOut{
		Listener: grpcmux.MuxListener{
			Tag:      "P2PGrpc",
			Listener: p,
		},
	}, nil
}
