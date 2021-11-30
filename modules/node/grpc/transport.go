package grpcsvc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

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
	m.Start(ctx, func(key string, err error) {
		dMap := map[string]interface{}{
			key: "Failed Err:" + err.Error(),
		}
		in.StManager.Report("RPC Listeners", status.Map(dMap))
	})

	stMp := make(map[string]interface{})
	stMapMtx := sync.Mutex{}
	stMapMtx.Lock()
	for _, v := range in.Listeners {
		stMp[v.Tag] = "Running"
	}
	stMapMtx.Unlock()
	in.StManager.Report("RPC Listeners", status.Map(stMp))

	lc.Append(fx.Hook{
		OnStop: func(c context.Context) error {
			defer func() {
				stMapMtx.Lock()
				for k := range stMp {
					stMp[k] = "Stopped"
				}
				stMapMtx.Unlock()
				in.StManager.Report("RPC Listeners", status.Map(stMp))
			}()
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
	log.Info("Starting TCP listener on port", portVal)
	listnr, err := net.Listen("tcp", fmt.Sprintf(":%d", portVal))
	if err != nil {
		log.Errorf("Failed starting TCP listener Err:%s", err.Error())
		return MuxListenerOut{}, err
	}
	return MuxListenerOut{
		Listener: grpcmux.MuxListener{
			Tag:      fmt.Sprintf("TCP Port %d", portVal),
			Listener: listnr,
		},
	}, nil
}

func NewP2PListener(
	h host.Host,
) (MuxListenerOut, error) {
	p, err := gostream.Listen(h, p2pgrpc.Protocol)
	if err != nil {
		return MuxListenerOut{}, err
	}
	log.Info("Started listener on P2P Host")
	return MuxListenerOut{
		Listener: grpcmux.MuxListener{
			Tag:      "P2PGrpc",
			Listener: p,
		},
	}, nil
}
