package ipfs

import (
	"go.uber.org/fx"
)

var P2PModule = fx.Options(
	fx.Provide(Identity),
	fx.Provide(Libp2p),
	fx.Provide(Pubsub),
	fx.Provide(NewSvcDiscovery),
	fx.Invoke(NewMDNSDiscovery),
)

var FilesModule = fx.Provide(NewNode)
