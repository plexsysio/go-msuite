package auth_test

import (
	"testing"

	ds "github.com/ipfs/go-datastore"
	syncds "github.com/ipfs/go-datastore/sync"
	store "github.com/plexsysio/gkvstore"
	dsstore "github.com/plexsysio/gkvstore-ipfsds"
	"github.com/plexsysio/go-msuite/modules/auth"
	"github.com/plexsysio/go-msuite/modules/config"
	jsonConf "github.com/plexsysio/go-msuite/modules/config/json"
)

type testRepo struct {
	st  store.Store
	cfg config.Config
	d   ds.Batching
}

func newTestRepo() *testRepo {
	cfg := jsonConf.DefaultConfig()
	bs := syncds.MutexWrap(ds.NewMapDatastore())
	st := dsstore.New(bs)
	return &testRepo{st: st, cfg: cfg, d: bs}
}

func (t *testRepo) Datastore() ds.Batching {
	return t.d
}

func (t *testRepo) Config() config.Config {
	return t.cfg
}

func (t *testRepo) Store() store.Store {
	return t.st
}

func (t *testRepo) SetConfig(c config.Config) error {
	t.cfg = c
	return nil
}

func (t *testRepo) Close() error {
	t.st.Close()
	t.d.Close()
	return nil
}

func TestNewAclManager(t *testing.T) {
	r := newTestRepo()
	defer r.Close()

	_, err := auth.NewAclManager(r)
	if err != nil {
		t.Fatal("Failed creating new acl manager", err.Error())
	}
}

func TestNewAclManagerWithAcls(t *testing.T) {
	r := newTestRepo()
	defer r.Close()

	r.Config().Set("ACL", map[string]string{
		"dummy": "invalidACL",
	})
	_, err := auth.NewAclManager(r)
	if err == nil {
		t.Fatal("Expected error while creating new acl manager")
	}

	r.Config().Set("ACL", map[string]string{
		"dummy": "admin",
	})
	_, err = auth.NewAclManager(r)
	if err != nil {
		t.Fatal("Failed creating new acl manager with ACLs", err.Error())
	}
}

func TestACLLifecycle(t *testing.T) {
	r := newTestRepo()
	defer r.Close()

	am, err := auth.NewAclManager(r)
	if err != nil {
		t.Fatal("Failed creating new acl manager", err.Error())
	}
	roles := am.Allowed("dummyresource")
	if len(roles) != 6 || roles[0] != auth.None {
		t.Fatalf("Invalid allowed role for no ACL %v", roles)
	}
	err = am.Configure("dummyresouce", "invalidACL")
	if err == nil {
		t.Fatal("Expected failure creating invalid ACL")
	}
	err = am.Configure("dummyresource", auth.Admin)
	if err != nil {
		t.Fatal("Failed creating new ACL", err.Error())
	}
	roles = am.Allowed("dummyresource")
	if len(roles) != 1 || roles[0] != auth.Admin {
		t.Fatalf("Invalid allowed role for no ACL %v", roles)
	}
	if am.Authorized("dummyresource", auth.AuthWrite) {
		t.Fatal("Authorized incorrect ACL")
	}
	err = am.Configure("dummyresource", auth.PublicWrite)
	if err != nil {
		t.Fatal("Failed creating new ACL", err.Error())
	}
	if !am.Authorized("dummyresource", auth.AuthWrite) {
		t.Fatal("Expected authorization for ACL", auth.AuthWrite)
	}
	if am.Authorized("dummyresource", auth.PublicRead) {
		t.Fatal("Authorized incorrect ACL")
	}
	roles = am.Allowed("dummyresource")
	if len(roles) != 4 {
		t.Fatal("Invalid allowed list", roles)
	}
	for _, rl := range roles {
		if rl == auth.PublicRead {
			t.Fatal("Invalid role in allowed list", rl)
		}
	}
	err = am.Delete("dummyresource")
	if err != nil {
		t.Fatal("Failed to delete ACL", err.Error())
	}
	if !am.Authorized("dummyresource", auth.AuthWrite) {
		t.Fatal("Expected authorization for ACL", auth.AuthWrite)
	}
}
