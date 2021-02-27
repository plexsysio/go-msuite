package ipfs

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(Identity),
	fx.Provide(Libp2p),
	fx.Provide(NewNode),
	fx.Provide(Pubsub),
	fx.Provide(NewSvcDiscovery),
	fx.Provide(NewAntsDB),
	fx.Invoke(NewMDNSDiscovery),
)
