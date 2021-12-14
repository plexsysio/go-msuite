package inmem_test

import (
	"reflect"
	"testing"

	jsonConf "github.com/plexsysio/go-msuite/modules/config/json"
	"github.com/plexsysio/go-msuite/modules/repo/inmem"
)

func TestInmemRepo(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		r, err := inmem.CreateOrOpen(jsonConf.DefaultConfig())
		if err != nil {
			t.Fatal(err)
		}

		if r.Datastore() == nil {
			t.Fatal("datastore not found")
		}

		if r.Store() == nil {
			t.Fatal("store not found")
		}

		if r.Status().(string) != "In-mem repository" {
			t.Fatal("invalid status")
		}

		err = r.Close()
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("create with config", func(t *testing.T) {
		cfg := jsonConf.DefaultConfig()
		cfg.Set("Dummy1", "Val1")
		cfg.Set("Dummy2", "Val2")

		r, err := inmem.CreateOrOpen(cfg)
		if err != nil {
			t.Fatal(err)
		}

		cfg2 := r.Config()
		if !reflect.DeepEqual(cfg, cfg2) {
			t.Fatal("expected config to be the same")
		}

		cfg3 := jsonConf.DefaultConfig()
		err = r.SetConfig(cfg3)
		if err != nil {
			t.Fatal(err)
		}

		cfg4 := r.Config()
		if !reflect.DeepEqual(cfg4, cfg3) {
			t.Fatal("expected config to be the same")
		}

		if reflect.DeepEqual(cfg4, cfg) {
			t.Fatal("expected config to be different")
		}

		if r.Datastore() == nil {
			t.Fatal("datastore not found")
		}

		if r.Store() == nil {
			t.Fatal("store not found")
		}

		if r.Status().(string) != "In-mem repository" {
			t.Fatal("invalid status")
		}

		err = r.Close()
		if err != nil {
			t.Fatal(err)
		}
	})
}
