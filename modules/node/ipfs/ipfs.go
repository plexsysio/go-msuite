package ipfs

import (
	"go.uber.org/fx"
)

var P2PModule = fx.Options(
	fx.Provide(Identity),
	fx.Provide(fx.Annotate(Libp2p, fx.ResultTags(`name:"mainHost"`, ``, ``))),
	fx.Provide(fx.Annotate(LocalDialer, fx.ResultTags(`name:"localDialer"`))),
	fx.Provide(fx.Annotate(Pubsub, fx.ParamTags(``, `name:"mainHost"`))),
	fx.Provide(NewSvcDiscovery),
	fx.Invoke(fx.Annotate(NewMDNSDiscovery, fx.ParamTags(``, `name:"mainHost"`))),
	fx.Invoke(fx.Annotate(NewP2PReporter, fx.ParamTags(`name:"mainHost"`, ``))),
	fx.Invoke(fx.Annotate(Bootstrapper, fx.ParamTags(``, ``, ``, `name:"mainHost"`))),
)

var FilesModule = fx.Provide(fx.Annotate(NewNode, fx.ParamTags(``, `name:"mainHost"`, ``, ``)))
