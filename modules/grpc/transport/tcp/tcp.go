package tcp

import (
	"context"
	"errors"
	"fmt"
	"github.com/aloknerurkar/go-msuite/modules/config"
	logger "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"
	"net"
)

var log = logger.Logger("transport/tcp")

func NewTCPListener(lc fx.Lifecycle, conf config.Config) (net.Listener, error) {
	var portVal int
	ok := conf.Get("TCPPort", &portVal)
	if !ok {
		return nil, errors.New("Port absent")
	}
	log.Info("Starting TCP listener on port", portVal)
	listnr, err := net.Listen("tcp", fmt.Sprintf(":%d", portVal))
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping listener")
			listnr.Close()
			return nil
		},
	})
	return listnr, nil
}
