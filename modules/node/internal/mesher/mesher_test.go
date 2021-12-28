package mesher_test

import (
	"context"
	"testing"
	"time"

	logger "github.com/ipfs/go-log/v2"
	bhost "github.com/libp2p/go-libp2p-blankhost"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peerstore"
	swarmt "github.com/libp2p/go-libp2p-swarm/testing"
	"github.com/plexsysio/go-msuite/modules/node/internal/mesher"
	"github.com/plexsysio/go-msuite/modules/protocols"
	"github.com/plexsysio/taskmanager"
)

// Tests the mesher protocol. If there is one bootstrap host, all nodes will
// be able to discover each other
func TestMesher(t *testing.T) {

	_ = logger.SetLogLevel("proto/mesher", "Debug")

	h1 := bhost.NewBlankHost(swarmt.GenSwarm(t, swarmt.OptDisableQUIC))
	h2 := bhost.NewBlankHost(swarmt.GenSwarm(t, swarmt.OptDisableQUIC))
	h3 := bhost.NewBlankHost(swarmt.GenSwarm(t, swarmt.OptDisableQUIC))
	h4 := bhost.NewBlankHost(swarmt.GenSwarm(t, swarmt.OptDisableQUIC))

	tm := taskmanager.New(0, 10, time.Second)

	t.Cleanup(func() {
		h1.Close()
		h2.Close()
		h3.Close()
		h4.Close()
		tm.Stop()
	})

	svc1 := protocols.New(h1)
	svc2 := protocols.New(h2)
	svc3 := protocols.New(h3)
	svc4 := protocols.New(h4)

	err := mesher.New(svc1, h1, tm)
	if err != nil {
		t.Fatal(err)
	}
	err = mesher.New(svc2, h2, tm)
	if err != nil {
		t.Fatal(err)
	}
	err = mesher.New(svc3, h3, tm)
	if err != nil {
		t.Fatal(err)
	}
	err = mesher.New(svc4, h4, tm)
	if err != nil {
		t.Fatal(err)
	}

	connectHosts(t, h1, h2)
	connectHosts(t, h1, h3)
	connectHosts(t, h1, h4)

	started := time.Now()
	for {
		time.Sleep(time.Second)

		complete := true
		for _, h := range []host.Host{h1, h2, h3, h4} {
			if len(h.Network().Peers()) != 3 {
				complete = false
				break
			}
		}

		if complete {
			return
		}

		if !complete && time.Since(started) > 5*time.Second {
			t.Fatal("waited 5 secs for peers to discover each other")
		}
	}
}

func connectHosts(t *testing.T, a, b host.Host) {
	t.Helper()

	ainfo := a.Peerstore().PeerInfo(a.ID())
	binfo := b.Peerstore().PeerInfo(b.ID())

	err := b.Connect(context.Background(), ainfo)
	if err != nil {
		t.Fatal(err)
	}

	// Add addresses on both sides
	a.Peerstore().AddAddrs(binfo.ID, binfo.Addrs, peerstore.PermanentAddrTTL)
	b.Peerstore().AddAddrs(ainfo.ID, ainfo.Addrs, peerstore.PermanentAddrTTL)
}
