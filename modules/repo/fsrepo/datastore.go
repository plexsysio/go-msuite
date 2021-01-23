package fsrepo

import (
	"github.com/aloknerurkar/go-msuite/modules/config"
	ds "github.com/ipfs/go-datastore"
)

type DatastoreCfg interface {
	Type() string
	AdditionalCfg() map[string]interface{}
}

func openDatastoreFromCfg(c config.Config) (ds.Batching, error) {
	return nil, nil
}
