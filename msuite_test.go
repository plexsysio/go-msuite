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
	_ = logger.SetLogLevel("*", "Debug")
	os.Exit(m.Run())
}

func MustRepo(t *testing.T, m core.Service, exists bool) {
	t.Helper()

	r := m.Repo()
	if r == nil && !exists {
		t.Fatal("Expected error accessing repo")
	}
}

func MustNode(t *testing.T, m core.Service, exists bool) {
	t.Helper()

	_, err := m.Node()
	if err == nil && !exists {
		t.Fatal("Expected error accessing Node")
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

func MustTM(t *testing.T, m core.Service, exists bool) {
	t.Helper()

	_, err := m.TM()
	if err == nil && !exists {
		t.Fatal("Expected error accessing TM")
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

func MustJWT(t *testing.T, m core.Service, exists bool) {
	t.Helper()

	_, err := m.Auth().JWT()
	if err == nil && !exists {
		t.Fatal("Expected error accessing JWT")
	}
}

func MustACL(t *testing.T, m core.Service, exists bool) {
	t.Helper()

	_, err := m.Auth().ACL()
	if err == nil && !exists {
		t.Fatal("Expected error accessing ACL")
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
	app, err := msuite.New(msuite.WithRepositoryRoot("tmp"))
	if err != nil {
		t.Fatal("Failed creating new msuite instance", err)
	}

	MustRepo(t, app, true)
	MustNode(t, app, false)
	MustGRPC(t, app, false)
	MustHTTP(t, app, false)
	MustTM(t, app, false)
	MustLocker(t, app, false)
	MustEvents(t, app, false)
	MustJWT(t, app, false)
	MustACL(t, app, false)
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
		msuite.WithTaskManager(5, 100),
	)
	if err != nil {
		t.Fatal("Failed creating new msuite instance", err)
	}

	MustRepo(t, app, true)
	MustTM(t, app, true)
	MustNode(t, app, false)
	MustGRPC(t, app, false)
	MustHTTP(t, app, false)
	MustLocker(t, app, false)
	MustEvents(t, app, false)
	MustJWT(t, app, false)
	MustACL(t, app, false)
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
		msuite.WithP2PPort(10000),
	)
	if err != nil {
		t.Fatal("Failed creating new msuite instance", err)
	}

	MustRepo(t, app, true)
	MustTM(t, app, true)
	MustNode(t, app, true)
	MustEvents(t, app, true)
	MustSharedStorage(t, app, true)
	MustGRPC(t, app, false)
	MustHTTP(t, app, false)
	MustLocker(t, app, false)
	MustJWT(t, app, false)
	MustACL(t, app, false)
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
		msuite.WithRepositoryRoot("tmp3"),
		msuite.WithHTTP(10000),
	)
	if err != nil {
		t.Fatal("Failed creating new msuite instance", err)
	}

	MustRepo(t, app, true)
	MustTM(t, app, true)
	MustHTTP(t, app, true)
	MustNode(t, app, false)
	MustGRPC(t, app, false)
	MustLocker(t, app, false)
	MustEvents(t, app, false)
	MustJWT(t, app, false)
	MustACL(t, app, false)
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
		msuite.WithP2PPort(10000),
		msuite.WithFiles(),
		msuite.WithGRPC(),
		msuite.WithGRPCTCPListener(10001),
		msuite.WithLocker("inmem", nil),
		msuite.WithJWT("dummysecret"),
		msuite.WithServiceACL(map[string]string{
			"dummyresource": "admin",
		}),
	)
	if err != nil {
		t.Fatal("Failed creating new msuite instance", err)
	}

	MustRepo(t, app, true)
	MustTM(t, app, true)
	MustNode(t, app, true)
	MustGRPC(t, app, true)
	MustLocker(t, app, true)
	MustEvents(t, app, true)
	MustJWT(t, app, true)
	MustACL(t, app, true)
	MustHTTP(t, app, false)
	MustSharedStorage(t, app, true)
	MustFiles(t, app, true)

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
		msuite.WithServiceName("test"),
		msuite.WithP2PPrivateKey(sk),
		msuite.WithRepositoryRoot("tmp"),
		msuite.WithP2PPort(10000),
	)
	if err != nil {
		t.Fatal("Failed creating new msuite instance", err)
	}

	MustRepo(t, app, true)
	MustTM(t, app, true)
	MustNode(t, app, true)
	MustEvents(t, app, true)
	MustSharedStorage(t, app, true)
	MustGRPC(t, app, false)
	MustLocker(t, app, false)
	MustJWT(t, app, false)
	MustACL(t, app, false)
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

	nd, _ := app.Node()
	if nd.P2P().Host().ID().Pretty() != idCfg {
		t.Fatal("incorrect id in P2P host expected", idCfg, nd.P2P().Host().ID().Pretty())
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
		msuite.WithServiceName("test"),
		msuite.WithRepositoryRoot("tmp5"),
		msuite.WithGRPCTCPListener(10000),
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

	initCalled := false

	app, err = msuite.New(
		msuite.WithServiceName("test"),
		msuite.WithGRPCTCPListener(10000),
		msuite.WithHTTP(10001),
		msuite.WithPrometheus(true),
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

	MustRepo(t, app, true)
	MustTM(t, app, true)
	MustGRPC(t, app, true)
	MustHTTP(t, app, true)
	MustNode(t, app, false)
	MustEvents(t, app, false)
	MustSharedStorage(t, app, false)
	MustLocker(t, app, false)
	MustJWT(t, app, false)
	MustACL(t, app, false)

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
