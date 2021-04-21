package msuite

import (
	"context"
	logger "github.com/ipfs/go-log/v2"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	logger.SetLogLevel("*", "Error")
	os.Exit(m.Run())
}

func TestBasicNew(t *testing.T) {
	defer os.RemoveAll("tmp")
	app, err := New(WithRepositoryRoot("tmp"))
	if err != nil {
		t.Fatal("Failed creating new msuite instance", err)
	}
	r := app.Repo()
	if r == nil {
		t.Fatal("Failed accessing repo")
	}
	_, err = app.Node()
	if err == nil {
		t.Fatal("Expected error accessing Node")
	}
	_, err = app.GRPC()
	if err == nil {
		t.Fatal("Expected error accessing GRPC")
	}
	_, err = app.TM()
	if err == nil {
		t.Fatal("Expected error accessing TM")
	}
	_, err = app.HTTP()
	if err == nil {
		t.Fatal("Expected error accessing HTTP")
	}
	_, err = app.Locker()
	if err == nil {
		t.Fatal("Expected error accessing Locker")
	}
	_, err = app.Events()
	if err == nil {
		t.Fatal("Expected error accessing Events")
	}
	_, err = app.Auth().JWT()
	if err == nil {
		t.Fatal("Expected error accessing JWT")
	}
	_, err = app.Auth().ACL()
	if err == nil {
		t.Fatal("Expected error accessing ACL")
	}
	err = app.Start(context.Background())
	if err != nil {
		t.Fatal("Failed starting app", err.Error())
	}
	err = app.Stop(context.Background())
	if err != nil {
		t.Fatal("Failed stopping app", err.Error())
	}
}

func TestTM(t *testing.T) {
	defer os.RemoveAll("tmp1")
	app, err := New(
		WithRepositoryRoot("tmp1"),
		WithTaskManager(5),
	)
	if err != nil {
		t.Fatal("Failed creating new msuite instance", err)
	}
	r := app.Repo()
	if r == nil {
		t.Fatal("Failed accessing repo")
	}
	_, err = app.TM()
	if err != nil {
		t.Fatal("Failed accessing TM", err.Error())
	}
	_, err = app.Node()
	if err == nil {
		t.Fatal("Expected error accessing Node")
	}
	_, err = app.GRPC()
	if err == nil {
		t.Fatal("Expected error accessing GRPC")
	}
	_, err = app.HTTP()
	if err == nil {
		t.Fatal("Expected error accessing HTTP")
	}
	_, err = app.Locker()
	if err == nil {
		t.Fatal("Expected error accessing Locker")
	}
	_, err = app.Events()
	if err == nil {
		t.Fatal("Expected error accessing Events")
	}
	_, err = app.Auth().JWT()
	if err == nil {
		t.Fatal("Expected error accessing JWT")
	}
	_, err = app.Auth().ACL()
	if err == nil {
		t.Fatal("Expected error accessing ACL")
	}
	err = app.Start(context.Background())
	if err != nil {
		t.Fatal("Failed starting app", err.Error())
	}
	err = app.Stop(context.Background())
	if err != nil {
		t.Fatal("Failed stopping app", err.Error())
	}
}

func TestNode(t *testing.T) {
	defer os.RemoveAll("tmp2")
	app, err := New(
		WithRepositoryRoot("tmp2"),
		WithP2PPort(10000),
	)
	if err != nil {
		t.Fatal("Failed creating new msuite instance", err)
	}
	r := app.Repo()
	if r == nil {
		t.Fatal("Failed accessing repo")
	}
	_, err = app.TM()
	if err != nil {
		t.Fatal("Failed accessing TM", err.Error())
	}
	_, err = app.Node()
	if err != nil {
		t.Fatal("Failed accessing Node", err.Error())
	}
	_, err = app.GRPC()
	if err != nil {
		t.Fatal("Failed accessing GRPC", err.Error())
	}
	_, err = app.HTTP()
	if err == nil {
		t.Fatal("Expected error accessing HTTP")
	}
	_, err = app.Locker()
	if err == nil {
		t.Fatal("Expected error accessing Locker")
	}
	_, err = app.Events()
	if err != nil {
		t.Fatal("Failed accessing Events", err.Error())
	}
	_, err = app.Auth().JWT()
	if err == nil {
		t.Fatal("Expected error accessing JWT")
	}
	_, err = app.Auth().ACL()
	if err == nil {
		t.Fatal("Expected error accessing ACL")
	}
	err = app.Start(context.Background())
	if err != nil {
		t.Fatal("Failed starting app", err.Error())
	}
	<-time.After(time.Second * 3)
	err = app.Stop(context.Background())
	if err != nil {
		t.Fatal("Failed stopping app", err.Error())
	}
}

func TestHTTP(t *testing.T) {
	defer os.RemoveAll("tmp3")
	app, err := New(
		WithRepositoryRoot("tmp3"),
		WithHTTP(10000),
	)
	if err != nil {
		t.Fatal("Failed creating new msuite instance", err)
	}
	r := app.Repo()
	if r == nil {
		t.Fatal("Failed accessing repo")
	}
	_, err = app.Node()
	if err == nil {
		t.Fatal("Expected error accessing Node")
	}
	_, err = app.GRPC()
	if err == nil {
		t.Fatal("Expected error accessing GRPC")
	}
	_, err = app.TM()
	if err != nil {
		t.Fatal("Failed accessing TM", err.Error())
	}
	_, err = app.HTTP()
	if err != nil {
		t.Fatal("Failed accessing HTTP", err.Error())
	}
	_, err = app.Locker()
	if err == nil {
		t.Fatal("Expected error accessing Locker")
	}
	_, err = app.Events()
	if err == nil {
		t.Fatal("Expected error accessing Events")
	}
	_, err = app.Auth().JWT()
	if err == nil {
		t.Fatal("Expected error accessing JWT")
	}
	_, err = app.Auth().ACL()
	if err == nil {
		t.Fatal("Expected error accessing ACL")
	}
	err = app.Start(context.Background())
	if err != nil {
		t.Fatal("Failed starting app", err.Error())
	}
	<-time.After(time.Second * 3)
	err = app.Stop(context.Background())
	if err != nil {
		t.Fatal("Failed stopping app", err.Error())
	}
}

func TestLockerAuth(t *testing.T) {
	defer os.RemoveAll("tmp4")
	app, err := New(
		WithRepositoryRoot("tmp4"),
		WithP2PPort(10000),
		WithGRPCTCPListener(10001),
		WithLocker("inmem", nil),
		WithJWT("dummysecret"),
		WithServiceACL(map[string]string{
			"dummyresource": "admin",
		}),
	)
	if err != nil {
		t.Fatal("Failed creating new msuite instance", err)
	}
	r := app.Repo()
	if r == nil {
		t.Fatal("Failed accessing repo")
	}
	_, err = app.TM()
	if err != nil {
		t.Fatal("Failed accessing TM", err.Error())
	}
	_, err = app.Node()
	if err != nil {
		t.Fatal("Failed accessing Node", err.Error())
	}
	_, err = app.GRPC()
	if err != nil {
		t.Fatal("Failed accessing GRPC", err.Error())
	}
	_, err = app.HTTP()
	if err == nil {
		t.Fatal("Expected error accessing HTTP")
	}
	_, err = app.Locker()
	if err != nil {
		t.Fatal("Failed accessing Locker", err.Error())
	}
	_, err = app.Events()
	if err != nil {
		t.Fatal("Failed accessing Events", err.Error())
	}
	_, err = app.Auth().JWT()
	if err != nil {
		t.Fatal("Failed accessing JWT", err.Error())
	}
	_, err = app.Auth().ACL()
	if err != nil {
		t.Fatal("Failed accessing ACL", err.Error())
	}
	err = app.Start(context.Background())
	if err != nil {
		t.Fatal("Failed starting app", err.Error())
	}
	<-time.After(time.Second * 3)
	err = app.Stop(context.Background())
	if err != nil {
		t.Fatal("Failed stopping app", err.Error())
	}
}
