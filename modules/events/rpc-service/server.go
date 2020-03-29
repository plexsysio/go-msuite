package events_service

//go:generate protoc -I/Users/aloknerurkar/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --proto_path=../pb --go_out=plugins,paths=source_relative:../pb ../pb/events.proto

import (
	"context"
	"fmt"
	Events "github.com/aloknerurkar/go-msuite/modules/events/pb"
	logger "github.com/ipfs/go-log"
	"google.golang.org/grpc"
	"sync"
)

var log = logger.Logger("events")

type events struct {
	svc      string
	handlers sync.Map
}

func (e *events) Handle(ctx context.Context, in *Events.Item) (*Events.Response, error) {
	val, ok := e.handlers.Load(in.Topic)
	if ok {
		handlers := val.([]func([]byte) error)
		log.Infof("Found %d handlers for topic %s", len(handlers), in.Topic)
		errs := make([]error, 0)
		for _, handler := range handlers {
			err := handler(in.Msg)
			if err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) == len(handlers) {
			log.Errorf("Event %s failed for all handlers. Returning error.")
			return nil, errs[0]
		}
		if len(errs) > 0 {
			log.Warningf("%d Handlers failed for topic %s. Errs: %v. Not returning error",
				len(errs), in.Topic, errs)
		}
		return &Events.Response{Result: "success"}, nil
	}
	return nil, fmt.Errorf(
		"Service %s doesnt implement handler for event:%s", e.svc, in.Topic)
}

func (e *events) RegisterHandler(topic string, handleBytes func(in []byte) error) {
	val, ok := e.handlers.Load(topic)
	if ok {
		hdlrs := val.([]func([]byte) error)
		hdlrs = append(hdlrs, handleBytes)
		log.Infof("Added new handler for topic %s No. of Handlers:%d",
			topic, len(hdlrs))
		return
	}
	e.handlers.Store(topic, []func(in []byte) error{handleBytes})
	log.Infof("New topic %s. Handler registered", topic)
	return
}

type RawEventHandler interface {
	RegisterHandler(string, func([]byte) error)
}

func NewEventsServer(srv *grpc.Server, svc string) RawEventHandler {
	e := &events{
		svc: svc,
	}
	Events.RegisterEventsServer(srv, e)
	return e
}
