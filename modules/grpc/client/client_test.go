package grpcclient_test

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	bhost "github.com/libp2p/go-libp2p-blankhost"
	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	swarmt "github.com/libp2p/go-libp2p-swarm/testing"
	jsonConf "github.com/plexsysio/go-msuite/modules/config/json"
	grpcclient "github.com/plexsysio/go-msuite/modules/grpc/client"
	"github.com/plexsysio/go-msuite/modules/grpc/p2pgrpc"
	"github.com/plexsysio/taskmanager"
	"google.golang.org/grpc"
)

func TestStaticAddrs(t *testing.T) {

	l1, err := net.Listen("tcp", ":10081")
	if err != nil {
		t.Fatal(err)
	}

	l2, err := net.Listen("unix", "/tmp/sock")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		l1.Close()
		l2.Close()
		_ = os.RemoveAll("/tmp/sock")
	})

	cfg := jsonConf.DefaultConfig()
	cfg.Set("StaticAddresses", map[string]string{
		"svc1": "localhost:10081",
		"svc2": "/tmp/sock",
	})

	c := grpcclient.NewStaticClientService(cfg)

	// Security credentials must be used, otherwise insecure should be explicitly
	// added
	_, err = c.Get(context.TODO(), "svc1")
	if err == nil {
		t.Fatal("succeeded without any options")
	}

	conn, err := c.Get(context.TODO(), "svc1", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()

	conn, err = c.Get(context.TODO(), "svc2", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

type testDiscovery struct {
	adv  chan string
	addr peer.AddrInfo
}

func (t *testDiscovery) Advertise(_ context.Context, ns string, _ ...discovery.Option) (time.Duration, error) {
	t.adv <- ns
	return time.Second, nil
}

func (t *testDiscovery) FindPeers(_ context.Context, _ string, _ ...discovery.Option) (<-chan peer.AddrInfo, error) {
	res := make(chan peer.AddrInfo)
	go func() {
		res <- t.addr
		close(res)
	}()
	return res, nil
}

func TestP2PClient(t *testing.T) {

	h1 := bhost.NewBlankHost(swarmt.GenSwarm(t, swarmt.OptDisableQUIC))

	h2Fired, h3Fired := make(chan struct{}), make(chan struct{})

	h2 := bhost.NewBlankHost(swarmt.GenSwarm(t, swarmt.OptDisableQUIC))
	h2.SetStreamHandler(p2pgrpc.Protocol, func(s network.Stream) {
		close(h2Fired)
	})

	h3 := bhost.NewBlankHost(swarmt.GenSwarm(t, swarmt.OptDisableQUIC))
	h3.SetStreamHandler(p2pgrpc.Protocol, func(s network.Stream) {
		close(h3Fired)
	})

	t.Cleanup(func() {
		h1.Close()
		h2.Close()
		h3.Close()
	})

	cfg := jsonConf.DefaultConfig()
	cfg.Set("Services", []string{"svc1"})

	cs, err := grpcclient.NewP2PClientService(
		cfg,
		&testDiscovery{addr: h3.Peerstore().PeerInfo(h3.ID())}, // discovery provide h3 for svc2
		h1, // local dialer
		h2, // local host with svc1
	)
	if err != nil {
		t.Fatal(err)
	}

	conn, err := cs.Get(context.TODO(), "svc1", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	<-h2Fired
	conn.Close()

	conn, err = cs.Get(context.TODO(), "svc2", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	<-h3Fired
	conn.Close()
}

func TestP2PAdvertiser(t *testing.T) {
	cfg := jsonConf.DefaultConfig()
	cfg.Set("Services", []string{"svc1", "svc2"})

	tm := taskmanager.New(0, 2, time.Second)
	t.Cleanup(func() {
		tm.Stop()
	})

	adv := make(chan string)
	d := &testDiscovery{adv: adv}
	err := grpcclient.NewP2PClientAdvertiser(cfg, d, tm)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)
	s := <-adv
	if s != "svc1" {
		t.Fatal("incorrect advertisement", s)
	}

	time.Sleep(100 * time.Millisecond)
	s = <-adv
	if s != "svc2" {
		t.Fatal("incorrect advertisement", s)
	}
}
