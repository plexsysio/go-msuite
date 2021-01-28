package fsrepo

import (
	"github.com/aloknerurkar/go-msuite/modules/config/json"
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	defer func() {
		os.RemoveAll(".testrepo")
		os.RemoveAll(".testrepo2")
	}()
	cfg := jsonConf.DefaultConfig()

	err := Init(".testrepo", cfg)
	if err != nil {
		t.Fatal("Failed initializing repo", err)
	}

	err = Init(".testrepo", cfg)
	if err == nil {
		t.Fatal("Should not be able to initialize already initialized repo")
	}

	err = Init(".testrepo2", cfg)
	if err != nil {
		t.Fatal("Failed to initialize repo with different path", err)
	}
}

func TestOpen(t *testing.T) {
	defer func() {
		os.RemoveAll(".testrepo")
	}()
	_, err := Open(".testrepo")
	if err == nil {
		t.Fatal("Able to open repo which is not initialized")
	}
	err = Init(".testrepo", jsonConf.DefaultConfig())
	if err != nil {
		t.Fatal("Failed to initialize repo", err)
	}
	r, err := Open(".testrepo")
	if err != nil {
		t.Fatal("Failed to open initialized repo", err)
	}
	r2, err := Open(".testrepo")
	if err != nil {
		t.Fatal("Unable to open already opened repo", err)
	}
	if r2 != r {
		t.Fatal("Newly opened repo doesnt match already open one")
	}
	if opener.refCnt != 2 {
		t.Fatal("RefCnt is incorrect", opener)
	}
	err = r.Close()
	if err != nil {
		t.Fatal("Failed closing repo", err)
	}
	if opener.refCnt != 1 {
		t.Fatal("RefCnt is incorrect", opener)
	}
	err = r2.Close()
	if err != nil {
		t.Fatal("Failed closing repo", err)
	}
	if opener.refCnt != 0 {
		t.Fatal("RefCnt is incorrect", opener)
	}
	if opener.active.kvStore != nil || opener.active.rootDS != nil {
		t.Fatal("Close did not clear internal stores")
	}
}
