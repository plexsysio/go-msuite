package repo

import (
	"io"

	ds "github.com/ipfs/go-datastore"
	store "github.com/plexsysio/gkvstore"
	"github.com/plexsysio/go-msuite/modules/config"
)

type Repo interface {
	Config() config.Config
	SetConfig(config.Config) error

	Datastore() ds.Batching

	Store() store.Store

	Status() interface{}
	io.Closer
}
