package repo

import (
	"github.com/SWRMLabs/ss-store"
	"github.com/aloknerurkar/go-msuite/modules/config"
	ds "github.com/ipfs/go-datastore"
	"io"
)

type Repo interface {
	Config() config.Config
	SetConfig(config.Config) error

	Datastore() ds.Batching

	Store() store.Store

	io.Closer
}
