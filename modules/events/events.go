package events

import (
	svc "github.com/aloknerurkar/go-msuite/modules/events/rpc-service"
	"github.com/golang/protobuf/proto"
	"go.uber.org/fx"
)

type PbFactory func() proto.Message
type HandlePb func(proto.Message) error

type PbEventBus interface {
	RegisterPbHandler(string, PbFactory, HandlePb)
}

var Bus = fx.Options(
	fx.Provide(svc.NewEventsServer),
	fx.Provide(NewPbEventBus),
)

type pbEventBus struct {
	r svc.RawEventHandler
}

func NewPbEventBus(rh svc.RawEventHandler) PbEventBus {
	return &pbEventBus{r: rh}
}

func (p *pbEventBus) RegisterPbHandler(
	topic string,
	factory PbFactory,
	handle HandlePb,
) {
	p.r.RegisterHandler(topic, func(in []byte) error {
		it := factory()
		err := proto.Unmarshal(in, it)
		if err != nil {
			return err
		}
		return handle(it)
	})
}
