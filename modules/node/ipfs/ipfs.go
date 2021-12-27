package ipfs

import (
	"go.uber.org/fx"
)

var P2PModule = fx.Options(
	fx.Provide(Identity),
	fx.Provide(Libp2p),
	fx.Provide(
		fx.Annotate(
			LocalDialer,
			fx.ResultTags(`name:"localDialer"`),
		),
	),
	fx.Provide(Pubsub),
	fx.Provide(NewSvcDiscovery),
	fx.Invoke(NewMDNSDiscovery),
	fx.Invoke(NewP2PReporter),
	fx.Invoke(Bootstrapper),
)

var FilesModule = fx.Provide(NewNode)
