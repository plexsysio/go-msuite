package fsrepo_test

import (
	"os"
	"testing"

	"github.com/plexsysio/go-msuite/modules/config/json"
	"github.com/plexsysio/go-msuite/modules/repo/fsrepo"
)

func TestInit(t *testing.T) {
	defer func() {
		os.RemoveAll(".testrepo")
		os.RemoveAll(".testrepo2")
	}()
	cfg := jsonConf.DefaultConfig()

	err := fsrepo.Init(".testrepo", cfg)
	if err != nil {
		t.Fatal("Failed initializing repo", err)
	}

	err = fsrepo.Init(".testrepo", cfg)
	if err == nil {
		t.Fatal("Should not be able to initialize already initialized repo")
	}

	err = fsrepo.Init(".testrepo2", cfg)
	if err != nil {
		t.Fatal("Failed to initialize repo with different path", err)
	}
}

func TestOpen(t *testing.T) {
	defer func() {
		os.RemoveAll(".testrepo")
	}()
	_, err := fsrepo.Open(".testrepo")
	if err == nil {
		t.Fatal("Able to open repo which is not initialized")
	}
	err = fsrepo.Init(".testrepo", jsonConf.DefaultConfig())
	if err != nil {
		t.Fatal("Failed to initialize repo", err)
	}
	r, err := fsrepo.Open(".testrepo")
	if err != nil {
		t.Fatal("Failed to open initialized repo", err)
	}
	r2, err := fsrepo.Open(".testrepo")
	if err != nil {
		t.Fatal("Unable to open already opened repo", err)
	}
	if r2 != r {
		t.Fatal("Newly opened repo doesnt match already open one")
	}
	if fsrepo.Opener.RefCnt != 2 {
		t.Fatal("RefCnt is incorrect", fsrepo.Opener.RefCnt)
	}
	err = r.Close()
	if err != nil {
		t.Fatal("Failed closing repo", err)
	}
	if fsrepo.Opener.RefCnt != 1 {
		t.Fatal("RefCnt is incorrect", fsrepo.Opener)
	}
	err = r2.Close()
	if err != nil {
		t.Fatal("Failed closing repo", err)
	}
	if fsrepo.Opener.RefCnt != 0 {
		t.Fatal("RefCnt is incorrect", fsrepo.Opener)
	}
	if fsrepo.Opener.Active != nil {
		t.Fatal("Close did not clear internal stores")
	}
}
