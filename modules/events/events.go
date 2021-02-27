package events

import (
	"context"
	"encoding/json"
	"github.com/StreamSpace/ss-taskmanager"
	logger "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-pubsub"
	"go.uber.org/fx"
	"sync"
)

var Module = fx.Provide(NewEventsSvc)

var log = logger.Logger("events")

type Event interface {
	Message
	Topic() string
}

type Message interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

type Factory func() Event

type Handle func(Event)

type Events interface {
	RegisterHandler(string, Factory, Handle)
	Broadcast(context.Context, Event) error
}

func NewEventsSvc(ps *pubsub.PubSub, tm *taskmanager.TaskManager) (Events, error) {
	eTopic, err := ps.Join("events")
	if err != nil {
		log.Errorf("Failed joining events channel Err:%s", err.Error())
		return nil, err
	}
	sub, err := eTopic.Subscribe()
	if err != nil {
		log.Errorf("Failed subscribing events channel Err:%s", err.Error())
		return nil, err
	}
	eSvc := &eventsImpl{
		topic:     eTopic,
		handlrMap: sync.Map{},
	}
	l := &eventsListener{
		sub:  sub,
		impl: eSvc,
	}
	tm.GoWork(l)
	return eSvc, nil
}

type eventMsg struct {
	Topic string
	Msg   []byte
}

func (e *eventMsg) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

func (e *eventMsg) Unmarshal(buf []byte) error {
	return json.Unmarshal(buf, e)
}

type eventsImpl struct {
	topic     *pubsub.Topic
	handlrMap sync.Map
}

type evHandler struct {
	factory Factory
	handle  Handle
}

func (p *eventsImpl) RegisterHandler(topic string, factory Factory, handle Handle) {
	hdlrs, ok := p.handlrMap.Load(topic)
	if !ok {
		hdlrs = []*evHandler{}
	}
	hdlrs = append(hdlrs.([]*evHandler), &evHandler{factory: factory, handle: handle})
	p.handlrMap.Store(topic, hdlrs)
	log.Infof("Registered new handler Topic: %s No. of Handlers: %s",
		topic, len(hdlrs.([]*evHandler)))
}

func (p *eventsImpl) Broadcast(ctx context.Context, e Event) error {
	buf, err := e.Marshal()
	if err != nil {
		log.Errorf("Failed marshaling event body Err:%s", err.Error())
		return err
	}
	ev := &eventMsg{
		Topic: e.Topic(),
		Msg:   buf,
	}
	bufToSend, err := ev.Marshal()
	if err != nil {
		log.Errorf("Failed marshaling event msg Err:%s", err.Error())
		return err
	}
	return p.topic.Publish(ctx, bufToSend)
}

type eventsListener struct {
	sub  *pubsub.Subscription
	impl *eventsImpl
}

func (e *eventsListener) Name() string {
	return "EventsListener"
}

func (e *eventsListener) Execute(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			log.Info("Stopping event listener")
			return nil
		default:
		}
		msg, err := e.sub.Next(ctx)
		if err != nil {
			log.Errorf("Failed getting next pubsub msg Err:%s", err.Error())
			continue
		}
		ev := &eventMsg{}
		err = ev.Unmarshal(msg.Data)
		if err != nil {
			log.Errorf("Failed getting event msg Err:%s", err.Error())
			continue
		}
		hdlrs, ok := e.impl.handlrMap.Load(ev.Topic)
		if !ok {
			log.Warnf("No handlers registered for topic: %s", ev.Topic)
			continue
		}
		log.Infof("Handling topic %s No. of handlers: %d", ev.Topic, len(hdlrs.([]*evHandler)))
		for _, h := range hdlrs.([]*evHandler) {
			it := h.factory()
			err := it.Unmarshal(ev.Msg)
			if err != nil {
				log.Errorf("Failed unmarshaling event body Err:%s", err.Error())
				continue
			}
			h.handle(it)
		}
	}
}
