package grpcServer

import (
	"context"
	"errors"
	"fmt"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"go.uber.org/fx"
	"net"
)

func NewTCPListener(lc fx.Lifecycle, conf config.Config) (net.Listener, error) {
	portVal, ok := conf.Get("grpc_port").(int32)
	if !ok {
		return nil, errors.New("Port absent")
	}
	log.Infof("Starting TCP listener on port %d", portVal)
	listnr, err := net.Listen("tcp", fmt.Sprintf(":%d", portVal))
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Debugf("Stopping TCP listener")
			listnr.Close()
			return nil
		},
	})
	return listnr, nil
}

var TCP = fx.Option(
	fx.Provide(NewTCPListener),
)
