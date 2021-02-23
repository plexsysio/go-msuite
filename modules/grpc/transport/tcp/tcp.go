package tcp

import (
	"errors"
	"fmt"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/aloknerurkar/go-msuite/modules/grpc/transport/mux"
	logger "github.com/ipfs/go-log/v2"
	"net"
)

var log = logger.Logger("transport/tcp")

func NewTCPListener(conf config.Config) (grpcmux.MuxListenerOut, error) {
	var portVal int
	ok := conf.Get("TCPPort", &portVal)
	if !ok {
		log.Error("TCPPort missing")
		return grpcmux.MuxListenerOut{}, errors.New("Port absent")
	}
	log.Info("Starting TCP listener on port", portVal)
	listnr, err := net.Listen("tcp", fmt.Sprintf(":%d", portVal))
	if err != nil {
		log.Errorf("Failed starting TCP listener Err:%s", err.Error())
		return grpcmux.MuxListenerOut{}, err
	}
	return grpcmux.MuxListenerOut{
		Listener: grpcmux.MuxListener{
			Tag:      "TCP",
			Listener: listnr,
		},
	}, nil
}
