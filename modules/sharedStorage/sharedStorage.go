package sharedStorage

import (
	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/routing"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	antsdb "github.com/plexsysio/ants-db"
	store "github.com/plexsysio/gkvstore"
)

type Callback interface {
	Put(string)
	Delete(string)
}

type Provider interface {
	SharedStorage(string, Callback) (store.Store, error)
}

func NewSharedStoreProvider(
	ds datastore.Batching,
	h host.Host,
	dht routing.Routing,
	ps *pubsub.PubSub,
) Provider {
	return &impl{
		h:   h,
		dht: dht,
		ds:  ds,
		ps:  ps,
	}
}

type impl struct {
	h   host.Host
	dht routing.Routing
	ps  *pubsub.PubSub
	ds  datastore.Batching
}

func (i *impl) SharedStorage(ns string, callback Callback) (store.Store, error) {
	opts := []antsdb.Option{antsdb.WithChannel(ns)}
	if callback != nil {
		opts = append(opts, antsdb.WithSubscriber(callback))
	}

	return antsdb.New(
		i.h,
		i.dht,
		i.ps,
		i.ds,
		opts...,
	)
}
