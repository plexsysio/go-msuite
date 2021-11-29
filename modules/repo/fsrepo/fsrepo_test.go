package fsrepo_test

import (
	"os"
	"reflect"
	"testing"

	jsonConf "github.com/plexsysio/go-msuite/modules/config/json"
	"github.com/plexsysio/go-msuite/modules/repo/fsrepo"
)

func TestInit(t *testing.T) {
	defer func() {
		os.RemoveAll(".testrepo")
		os.RemoveAll(".testrepo2")
	}()
	cfg := jsonConf.DefaultConfig()

	if fsrepo.IsInitialized(".testrepo") {
		t.Fatal("expected repo to not be initialized")
	}

	err := fsrepo.Init(".testrepo", cfg)
	if err != nil {
		t.Fatal("Failed initializing repo", err)
	}

	if !fsrepo.IsInitialized(".testrepo") {
		t.Fatal("expected repo to be initialized")
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
	if fsrepo.Opener.ActiveMap[".testrepo"].RefCnt != 2 {
		t.Fatal("RefCnt is incorrect", fsrepo.Opener.ActiveMap[".testrepo"].RefCnt)
	}
	err = r.Close()
	if err != nil {
		t.Fatal("Failed closing repo", err)
	}
	if fsrepo.Opener.ActiveMap[".testrepo"].RefCnt != 1 {
		t.Fatal("RefCnt is incorrect", fsrepo.Opener)
	}
	err = r2.Close()
	if err != nil {
		t.Fatal("Failed closing repo", err)
	}
	if _, found := fsrepo.Opener.ActiveMap[".testrepo"]; found {
		t.Fatal("Repo should not be present in active map", fsrepo.Opener)
	}
}

func TestRepo(t *testing.T) {
	defer func() {
		os.RemoveAll(".testrepo")
	}()

	cfg := jsonConf.DefaultConfig()

	_, err := fsrepo.CreateOrOpen(cfg)
	if err == nil {
		t.Fatal("able to create repo without path")
	}

	cfg.Set("RootPath", ".testrepo")
	r, err := fsrepo.CreateOrOpen(cfg)
	if err != nil {
		t.Fatal("failed to create repo", err)
	}

	cfg2 := r.Config()
	if cfg2 == nil {
		t.Fatal("expected to find config")
	}

	if !reflect.DeepEqual(cfg, cfg2) {
		t.Fatal("expected to find same config")
	}

	s := r.Store()
	if s == nil {
		t.Fatal("expected to find store")
	}

	ds := r.Datastore()
	if ds == nil {
		t.Fatal("expected to find datastore")
	}

	mds, ok := ds.(fsrepo.Datastore)
	if !ok {
		t.Fatal("expected to find mounted datastore")
	}

	mntInfo, err := mds.Mounts()
	if err != nil {
		t.Fatal(err)
	}

	if len(mntInfo) != 2 {
		t.Fatal("expected 2 mounts found", len(mntInfo))
	}

	for _, v := range mntInfo {
		if v.Path != ".testrepo/kv" && v.Path != ".testrepo/blocks" {
			t.Fatal("unexpected paths", v.Path)
		}
		if v.Prefix != "/" && v.Prefix != "/blocks" {
			t.Fatal("unexpected prefixes", v.Prefix)
		}
	}

	cfg.Set("TestAdd", "some value")
	err = r.SetConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}

	err = r.Close()
	if err != nil {
		t.Fatal(err)
	}

	r, err = fsrepo.CreateOrOpen(cfg)
	if err != nil {
		t.Fatal(err)
	}

	var testAdd string
	found := r.Config().Get("TestAdd", &testAdd)
	if !found || testAdd != "some value" {
		t.Fatal("unexpected config on reopen", found, testAdd)
	}

	err = r.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestInvalidDatastoreConfig(t *testing.T) {
	defer func() {
		os.RemoveAll(".testrepo")
	}()

	cfg := jsonConf.DefaultConfig()

	cfg.Set("RootPath", ".testrepo")

	// invalid datastore type
	cfg.Set("Mounts", map[string]interface{}{
		"invalid": "invalid",
	})

	_, err := fsrepo.CreateOrOpen(cfg)
	if err == nil {
		t.Fatal("able to create repo with incorrect config", cfg.String())
	}

	// invalid sub DS config
	cfg.Set("Mounts", map[string]interface{}{
		"level": "invalid",
	})

	_, err = fsrepo.CreateOrOpen(cfg)
	if err == nil {
		t.Fatal("able to create repo with incorrect config", cfg.String())
	}

	// sub DS config without prefix
	cfg.Set("Mounts", map[string]interface{}{
		"level": map[string]interface{}{
			"path": "kv",
		},
	})

	_, err = fsrepo.CreateOrOpen(cfg)
	if err == nil {
		t.Fatal("able to create repo with incorrect config", cfg.String())
	}

	// one correct one incorrect
	cfg.Set("Mounts", map[string]interface{}{
		"level": map[string]interface{}{
			"path":   "kv",
			"prefix": "/",
		},
		"invalid": "invalid",
	})

	_, err = fsrepo.CreateOrOpen(cfg)
	if err == nil {
		t.Fatal("able to create repo with incorrect config", cfg.String())
	}
}
