package net

import (
	"errors"
	"fmt"
	"github.com/aloknerurkar/go-msuite/modules/config"
	logger "github.com/ipfs/go-log"
	"go.uber.org/fx"
	"net"
)

var log = logger.Logger("net/tcp")

func NewTCPListener(conf config.Config) (net.Listener, error) {
	portVal, ok := conf.Get("grpc_port").(int32)
	if !ok {
		return nil, errors.New("Port absent")
	}

	log.Infof("Starting TCP listener on port %d", portVal)
	listnr, err := net.Listen("tcp", fmt.Sprintf(":%d", portVal))
	if err != nil {
		return nil, err
	}

	return listnr, nil
}

var TCP = fx.Option(
	fx.Provide(NewTCPListener),
)
