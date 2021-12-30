package mesher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	logger "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/plexsysio/go-msuite/modules/protocols"
	"github.com/plexsysio/taskmanager"
)

var log = logger.Logger("proto/mesher")

type peersList []peer.AddrInfo

func (p *peersList) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func (p *peersList) Unmarshal(buf []byte) error {
	return json.Unmarshal(buf, p)
}

type service struct {
	h    host.Host
	send protocols.Sender
}

func New(svc protocols.ProtocolsSvc, h host.Host, tm *taskmanager.TaskManager) error {
	s := &service{h: h}
	svc.Register(s)

	newPeerChan := make(chan peer.ID)

	notifier := &network.NotifyBundle{
		ConnectedF: func(_ network.Network, conn network.Conn) {
			newPeerChan <- conn.RemotePeer()
		},
	}

	_, err := tm.GoFunc(fmt.Sprintf("mesher broadcaster worker %s", h.ID()), func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case p := <-newPeerChan:
				err := s.BroadcastPeers(ctx, p)
				if err != nil {
					log.Warn("failed broadcasting peers", err)
				}
			}
		}
	})
	if err != nil {
		return err
	}

	h.Network().Notify(notifier)
	return nil
}

func (service) ID() protocol.ID { return protocol.ID("/msuite/mesher/1.0.0") }

func (service) ReqFactory() protocols.Request { return new(peersList) }

func (service) RespFactory() protocols.Response { return new(peersList) }

func (s *service) SetSender(sender protocols.Sender) { s.send = sender }

func (s *service) HandleMsg(req protocols.Request, p peer.ID) (protocols.Response, error) {
	peers, ok := req.(*peersList)
	if !ok {
		return nil, errors.New("incorrect msg received")
	}

	resp := s.checkAndAddPeers(*peers)
	return resp, nil
}

func (s *service) getPeersFor(pr peer.ID) *peersList {
	currentPeers := s.h.Network().Peers()

	req := new(peersList)
	for _, p := range currentPeers {
		if p == pr || p == s.h.ID() {
			continue
		}
		info := s.h.Peerstore().PeerInfo(p)
		if len(info.Addrs) > 0 {
			*req = append(*req, info)
		}
	}

	return req
}

func (s *service) checkAndAddPeers(peers []peer.AddrInfo) *peersList {
	successful := new(peersList)
	for _, p := range peers {
		if s.h.Network().Connectedness(p.ID) != network.Connected {
			err := s.h.Connect(context.Background(), p)
			if err != nil {
				log.Warn("could not connect to peer", p)
				continue
			}

			s.h.Peerstore().AddAddrs(p.ID, p.Addrs, peerstore.PermanentAddrTTL)
			log.Debug("connected to peer", p)
			*successful = append(*successful, p)
		} else {
			log.Debug("already connected to peer", p)
			*successful = append(*successful, p)
		}
	}
	return successful
}

func (s *service) BroadcastPeers(ctx context.Context, p peer.ID) error {
	req := s.getPeersFor(p)

	if len(*req) == 0 {
		log.Warn("no peers to broadcast")
		return nil
	}

	// TODO: Currently we get the successfully connected peer list here which can be used
	// to improve the peerstore by removing addresses which are not connectable
	_, err := s.send(ctx, p, req)
	if err != nil {
		return err
	}

	return nil
}
