package libp2p

import (
	logger "github.com/ipfs/go-log"
	protocol "github.com/libp2p/go-libp2p-core/protocol"
	"go.uber.org/fx"
)

var P2PGrpc = fx.Options(
	fx.Provide(NewP2PListener),
	fx.Provide(NewP2PDialer),
)

var P2P = fx.Options(
	fx.Provide(Libp2p),
	fx.Invoke(NewMDNSDiscovery),
)

const Protocol protocol.ID = "/grpc/1.0.0"

const Rendezvous string = "/msuite/node"

var log = logger.Logger("libp2p")
