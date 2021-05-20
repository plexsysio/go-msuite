package sharedStorage

import (
	"context"
	"github.com/SWRMLabs/ants-db"
	"github.com/SWRMLabs/ss-store"
	"github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	"github.com/ipfs/go-ipfs-blockstore"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"go.uber.org/fx"
)

var Module = fx.Provide(NewSharedStoreProvider)

type Callback interface {
	Put(string)
	Delete(string)
}

type Provider interface {
	SharedStorage(string, Callback) (store.Store, error)
}

func NewSharedStoreProvider(
	ds datastore.Batching,
	peer *ipfslite.Peer,
	ps *pubsub.PubSub,
) Provider {
	return &impl{
		p:  peer,
		ds: ds,
	}
}

type impl struct {
	p  *ipfslite.Peer
	ps *pubsub.PubSub
	ds datastore.Batching
}

func (i *impl) SharedStorage(ns string, callback Callback) (store.Store, error) {
	nsk := datastore.NewKey(ns)
	ds := namespace.Wrap(i.ds, nsk)
	bs := blockstore.NewBlockstore(ds)

	syncer := &nsSyncer{
		DAGService: merkledag.NewDAGService(blockservice.New(bs, i.p.Exchange())),
		bs:         bs,
	}

	opts := []antsdb.Option{antsdb.WithChannel(ns)}
	if callback != nil {
		opts = append(opts, antsdb.WithSubscriber(callback))
	}

	return antsdb.New(
		syncer,
		i.ps,
		i.ds,
		opts...,
	)
}

type nsSyncer struct {
	ipld.DAGService

	bs blockstore.Blockstore
}

func (n *nsSyncer) HasBlock(c cid.Cid) (bool, error) {
	return n.bs.Has(c)

}

func (n *nsSyncer) Session(ctx context.Context) ipld.NodeGetter {
	return merkledag.NewSession(ctx, n.DAGService)
}
