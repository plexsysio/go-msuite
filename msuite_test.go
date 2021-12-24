package msuite_test

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"testing"
	"time"

	logger "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/plexsysio/go-msuite"
	"github.com/plexsysio/go-msuite/core"
)

func TestMain(m *testing.M) {
	_ = logger.SetLogLevel("*", "Error")
	os.Exit(m.Run())
}

func MustP2P(t *testing.T, m core.Service, exists bool) {
	t.Helper()

	_, err := m.P2P()
	if err == nil && !exists {
		t.Fatal("Expected error accessing P2P")
	}
}

func MustGRPC(t *testing.T, m core.Service, exists bool) {
	t.Helper()

	_, err := m.GRPC()
	if err == nil && !exists {
		t.Fatal("Expected error accessing GRPC")
	}
}

func MustHTTP(t *testing.T, m core.Service, exists bool) {
	t.Helper()

	_, err := m.HTTP()
	if err == nil && !exists {
		t.Fatal("Expected error accessing HTTP")
	}
}

func MustLocker(t *testing.T, m core.Service, exists bool) {
	t.Helper()

	_, err := m.Locker()
	if err == nil && !exists {
		t.Fatal("Expected error accessing Locker")
	}
}

func MustEvents(t *testing.T, m core.Service, exists bool) {
	t.Helper()

	_, err := m.Events()
	if err == nil && !exists {
		t.Fatal("Expected error accessing Events")
	}
}

func MustAuth(t *testing.T, m core.Service, exists bool) {
	t.Helper()

	_, err := m.Auth()
	if err == nil && !exists {
		t.Fatal("Expected error accessing Auth")
	}
}

func MustSharedStorage(t *testing.T, m core.Service, exists bool) {
	t.Helper()

	st, err := m.SharedStorage("test", nil)
	if err == nil && !exists {
		t.Fatal("Expected error accessing SharedStorage")
	}
	if st != nil {
		st.Close()
	}
}

func MustFiles(t *testing.T, m core.Service, exists bool) {
	t.Helper()

	_, err := m.Files()
	if err == nil && !exists {
		t.Fatal("expected error accessing files")
	}
}

func TestBasicNew(t *testing.T) {
	defer os.RemoveAll("tmp")
	app, err := msuite.New(
		context.TODO(),
		msuite.WithRepositoryRoot("tmp"),
	)
	if err != nil {
		t.Fatal("Failed creating new msuite instance", err)
	}

	MustP2P(t, app, false)
	MustGRPC(t, app, false)
	MustHTTP(t, app, false)
	MustLocker(t, app, false)
	MustEvents(t, app, false)
	MustAuth(t, app, false)
	MustSharedStorage(t, app, false)

	err = app.Start(context.Background())
	if err != nil {
		t.Fatal("Failed starting app", err.Error())
	}
	time.Sleep(time.Millisecond * 100)
	err = app.Stop(context.Background())
	if err != nil {
		t.Fatal("Failed stopping app", err.Error())
	}
}

func TestTM(t *testing.T) {
	app, err := msuite.New(
		context.TODO(),
		msuite.WithTaskManager(5, 100),
		// Auth without P2P should initialize OK
		msuite.WithAuth("dummysecret"),
	)
	if err != nil {
		t.Fatal("Failed creating new msuite instance", err)
	}

	MustAuth(t, app, true)
	MustP2P(t, app, false)
	MustGRPC(t, app, false)
	MustHTTP(t, app, false)
	MustLocker(t, app, false)
	MustEvents(t, app, false)
	MustSharedStorage(t, app, false)

	err = app.Start(context.Background())
	if err != nil {
		t.Fatal("Failed starting app", err.Error())
	}
	time.Sleep(time.Millisecond * 100)
	err = app.Stop(context.Background())
	if err != nil {
		t.Fatal("Failed stopping app", err.Error())
	}
}

func TestNode(t *testing.T) {
	app, err := msuite.New(
		context.TODO(),
		msuite.WithP2P(10000),
	)
	if err != nil {
		t.Fatal("Failed creating new msuite instance", err)
	}

	MustP2P(t, app, true)
	MustEvents(t, app, true)
	MustSharedStorage(t, app, true)
	MustGRPC(t, app, false)
	MustHTTP(t, app, false)
	MustLocker(t, app, false)
	MustAuth(t, app, false)
	MustFiles(t, app, false)

	err = app.Start(context.Background())
	if err != nil {
		t.Fatal("Failed starting app", err.Error())
	}
	time.Sleep(time.Millisecond * 100)
	err = app.Stop(context.Background())
	if err != nil {
		t.Fatal("Failed stopping app", err.Error())
	}
}

func TestHTTP(t *testing.T) {
	defer os.RemoveAll("tmp3")
	app, err := msuite.New(
		context.TODO(),
		msuite.WithRepositoryRoot("tmp3"),
		msuite.WithHTTP(10000),
	)
	if err != nil {
		t.Fatal("Failed creating new msuite instance", err)
	}

	MustHTTP(t, app, true)
	MustP2P(t, app, false)
	MustGRPC(t, app, false)
	MustLocker(t, app, false)
	MustEvents(t, app, false)
	MustAuth(t, app, false)
	MustSharedStorage(t, app, false)

	err = app.Start(context.Background())
	if err != nil {
		t.Fatal("Failed starting app", err.Error())
	}
	time.Sleep(time.Millisecond * 100)
	err = app.Stop(context.Background())
	if err != nil {
		t.Fatal("Failed stopping app", err.Error())
	}
}

func TestGRPCLockerAuth(t *testing.T) {
	app, err := msuite.New(
		context.TODO(),
		msuite.WithP2P(10000),
		msuite.WithFiles(),
		msuite.WithGRPC("tcp", 10001),
		msuite.WithGRPC("p2p", nil),
		msuite.WithLocker("inmem", nil),
		msuite.WithAuth("dummysecret"),
		msuite.WithServiceACL(map[string]string{
			"dummyresource": "admin",
		}),
	)
	if err != nil {
		t.Fatal("Failed creating new msuite instance", err)
	}

	MustP2P(t, app, true)
	MustGRPC(t, app, true)
	MustLocker(t, app, true)
	MustEvents(t, app, true)
	MustAuth(t, app, true)
	MustSharedStorage(t, app, true)
	MustFiles(t, app, true)
	MustHTTP(t, app, false)

	err = app.Start(context.Background())
	if err != nil {
		t.Fatal("Failed starting app", err.Error())
	}
	time.Sleep(time.Millisecond * 100)
	err = app.Stop(context.Background())
	if err != nil {
		t.Fatal("Failed stopping app", err.Error())
	}
}

func TestPrivateKey(t *testing.T) {
	defer os.RemoveAll("tmp")

	sk, pk, err := crypto.GenerateKeyPair(crypto.Ed25519, 2048)
	if err != nil {
		t.Fatal(err)
	}

	skbytes, err := crypto.MarshalPrivateKey(sk)
	if err != nil {
		t.Fatal(err)
	}

	privKeyStr := base64.StdEncoding.EncodeToString(skbytes)

	id, err := peer.IDFromPublicKey(pk)
	if err != nil {
		t.Fatal(err)
	}

	app, err := msuite.New(
		context.TODO(),
		msuite.WithServiceName("test"),
		msuite.WithP2PPrivateKey(sk),
		msuite.WithRepositoryRoot("tmp"),
		msuite.WithP2P(10000),
	)
	if err != nil {
		t.Fatal("Failed creating new msuite instance", err)
	}

	MustP2P(t, app, true)
	MustEvents(t, app, true)
	MustSharedStorage(t, app, true)
	MustGRPC(t, app, false)
	MustLocker(t, app, false)
	MustAuth(t, app, false)
	MustHTTP(t, app, false)

	identity := map[string]interface{}{}

	found := app.Repo().Config().Get("Identity", &identity)
	if !found {
		t.Fatal("expected to find privkey in config")
	}

	privKeyCfg := identity["PrivKey"].(string)
	if privKeyCfg != privKeyStr {
		t.Fatal("expected privkey", privKeyStr, "found", privKeyCfg)
	}

	idCfg := identity["ID"].(string)
	if idCfg != id.Pretty() {
		t.Fatal("expected ID", id.Pretty(), "found", idCfg)
	}

	nd, _ := app.P2P()
	if nd.Host().ID().Pretty() != idCfg {
		t.Fatal("incorrect id in P2P host expected", idCfg, nd.Host().ID().Pretty())
	}

	err = app.Start(context.Background())
	if err != nil {
		t.Fatal("Failed starting app", err.Error())
	}
	time.Sleep(time.Millisecond * 100)
	err = app.Stop(context.Background())
	if err != nil {
		t.Fatal("Failed stopping app", err.Error())
	}
}

func TestServices(t *testing.T) {
	defer os.RemoveAll("tmp5")

	app, err := msuite.New(
		context.TODO(),
		msuite.WithServiceName("test"),
		msuite.WithRepositoryRoot("tmp5"),
		msuite.WithGRPC("tcp", 10000),
		msuite.WithStaticDiscovery(map[string]string{
			"svc1": "IP1",
			"svc2": "IP2",
		}),
		msuite.WithService("testErr", func(_ core.Service) error {
			return errors.New("dummy error")
		}),
	)
	if err == nil || app != nil {
		t.Fatal("Expected error while creating new msuite instance")
	}

	time.Sleep(time.Second)

	initCalled := false

	app, err = msuite.New(
		context.TODO(),
		msuite.WithServiceName("test"),
		msuite.WithGRPC("tcp", 10000),
		msuite.WithHTTP(10001),
		msuite.WithPrometheus(true),
		msuite.WithDebug(),
		msuite.WithStaticDiscovery(map[string]string{
			"svc1": "IP1",
			"svc2": "IP2",
		}),
		msuite.WithService("testErr", func(_ core.Service) error {
			initCalled = true
			return nil
		}),
	)
	if err != nil {
		t.Fatal("Failed creating new msuite instance", err)
	}

	if !initCalled {
		t.Fatal("service not initialized")
	}

	MustGRPC(t, app, true)
	MustHTTP(t, app, true)
	MustP2P(t, app, false)
	MustEvents(t, app, false)
	MustSharedStorage(t, app, false)
	MustLocker(t, app, false)
	MustAuth(t, app, false)

	err = app.Start(context.Background())
	if err != nil {
		t.Fatal("Failed starting app", err.Error())
	}
	time.Sleep(time.Millisecond * 100)
	err = app.Stop(context.Background())
	if err != nil {
		t.Fatal("Failed stopping app", err.Error())
	}
}
