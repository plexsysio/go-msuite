package ipfs

import (
	logger "github.com/ipfs/go-log"
	"go.uber.org/fx"
)

var P2P = fx.Options(
	fx.Provide(Libp2p),
	fx.Invoke(NewMDNSDiscovery),
)

var log = logger.Logger("libp2p")
