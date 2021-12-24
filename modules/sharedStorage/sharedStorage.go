package sharedStorage

import (
	"strings"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/routing"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	antsdb "github.com/plexsysio/ants-db"
	store "github.com/plexsysio/gkvstore"
	"github.com/plexsysio/go-msuite/modules/config"
)

const defaultRootNs = "msuite"

type Callback interface {
	Put(string)
	Delete(string)
}

type Provider interface {
	SharedStorage(string, Callback) (store.Store, error)
}

func NewSharedStoreProvider(
	cfg config.Config,
	ds datastore.Batching,
	h host.Host,
	dht routing.Routing,
	ps *pubsub.PubSub,
) (Provider, error) {

	var (
		rootNs string
		err    error
	)
	_ = cfg.Get("SharedStoreNs", &rootNs)
	if rootNs == "" {
		rootNs = defaultRootNs
	}

	i := &impl{
		callbacks: make(map[string][]Callback),
	}

	i.shStore, err = antsdb.New(
		h,
		dht,
		ps,
		ds,
		antsdb.WithChannel(rootNs),
		antsdb.WithSubscriber(i),
	)
	if err != nil {
		return nil, err
	}

	return i, nil
}

type impl struct {
	shStore   store.Store
	cbLock    sync.RWMutex
	callbacks map[string][]Callback
}

func (i *impl) Put(key string) {
	i.cbLock.RLock()
	defer i.cbLock.RUnlock()

	for k, cbs := range i.callbacks {
		if strings.HasPrefix(key, k) {
			for _, cb := range cbs {
				cb.Put(key)
			}
		}
	}
}

func (i *impl) Delete(key string) {
	i.cbLock.RLock()
	defer i.cbLock.RUnlock()

	for k, cbs := range i.callbacks {
		if strings.HasPrefix(key, k) {
			for _, cb := range cbs {
				cb.Delete(key)
			}
		}
	}
}

func (i *impl) SharedStorage(ns string, callback Callback) (store.Store, error) {
	if callback != nil {
		i.cbLock.Lock()
		defer i.cbLock.Unlock()

		cbs, ok := i.callbacks[ns]
		if !ok {
			cbs = []Callback{callback}
		} else {
			cbs = append(cbs, callback)
		}

		i.callbacks[ns] = cbs
	}

	return i.shStore, nil
}
