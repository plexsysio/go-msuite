package grpcsvc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/libp2p/go-libp2p-core/host"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"github.com/plexsysio/go-msuite/modules/config"
	"github.com/plexsysio/go-msuite/modules/diag/status"
	grpcmux "github.com/plexsysio/go-msuite/modules/grpc/mux"
	"github.com/plexsysio/go-msuite/modules/grpc/p2pgrpc"
	"github.com/plexsysio/taskmanager"
	"go.uber.org/fx"
)

type MuxListenerOut struct {
	fx.Out

	Listener grpcmux.MuxListener `group:"listener"`
}

type MuxIn struct {
	fx.In

	Listeners []grpcmux.MuxListener `group:"listener"`
	StManager status.Manager        `optional:"true"`
}

func NewMuxedListener(
	ctx context.Context,
	lc fx.Lifecycle,
	in MuxIn,
	tm *taskmanager.TaskManager,
) (*grpcmux.Mux, error) {
	m := grpcmux.New(ctx, in.Listeners, tm)
	in.StManager.AddReporter("RPC Listeners", m)

	lc.Append(fx.Hook{
		OnStart: func(c context.Context) error {
			return m.Start(c)
		},
		OnStop: func(c context.Context) error {
			log.Info("Stopping Muxed listeners")
			err := m.Close()
			if err != nil {
				log.Warn("Error on closing listeners", err.Error())
			}
			return nil
		},
	})
	return m, nil
}

func NewTCPListener(conf config.Config) (MuxListenerOut, error) {
	var portVal int
	ok := conf.Get("TCPPort", &portVal)
	if !ok {
		log.Error("TCPPort missing")
		return MuxListenerOut{}, errors.New("Port absent")
	}
	return MuxListenerOut{
		Listener: grpcmux.MuxListener{
			Tag: fmt.Sprintf("TCP Port %d", portVal),
			Start: func() (net.Listener, error) {
				return net.Listen("tcp", fmt.Sprintf(":%d", portVal))
			},
		},
	}, nil
}

func NewP2PListener(
	h host.Host,
) (MuxListenerOut, error) {
	return MuxListenerOut{
		Listener: grpcmux.MuxListener{
			Tag: "P2PGrpc",
			Start: func() (net.Listener, error) {
				return gostream.Listen(h, p2pgrpc.Protocol)
			},
		},
	}, nil
}

func NewUDSListener(conf config.Config) (MuxListenerOut, error) {
	var sock string
	ok := conf.Get("UDSocket", &sock)
	if !ok {
		log.Error("Unix socket missing")
		return MuxListenerOut{}, errors.New("socket absent")
	}
	return MuxListenerOut{
		Listener: grpcmux.MuxListener{
			Tag: fmt.Sprintf("UDS Sock %s", sock),
			Start: func() (net.Listener, error) {
				return net.Listen("unix", sock)
			},
		},
	}, nil
}
