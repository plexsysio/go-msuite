package inmem

import (
	"encoding/base64"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/plexsysio/gkvstore"
	ipfsdsStore "github.com/plexsysio/gkvstore-ipfsds"
	"github.com/plexsysio/go-msuite/modules/config"
	"github.com/plexsysio/go-msuite/modules/repo"
)

type inmemRepo struct {
	c  config.Config
	ds datastore.Batching
	st gkvstore.Store
}

func initIdentity(c config.Config) error {
	if c.Get("Identity", make(map[string]interface{})) {
		return nil
	}
	sk, pk, err := crypto.GenerateKeyPair(crypto.Ed25519, 2048)
	if err != nil {
		return err
	}
	skbytes, err := crypto.MarshalPrivateKey(sk)
	if err != nil {
		return err
	}
	ident := map[string]interface{}{}
	ident["PrivKey"] = base64.StdEncoding.EncodeToString(skbytes)

	id, err := peer.IDFromPublicKey(pk)
	if err != nil {
		return err
	}
	ident["ID"] = id.Pretty()
	c.Set("Identity", ident)
	return nil
}

func CreateOrOpen(c config.Config) (repo.Repo, error) {
	err := initIdentity(c)
	if err != nil {
		return nil, err
	}
	ds := datastore.NewMapDatastore()
	st := ipfsdsStore.New(namespace.Wrap(ds, datastore.NewKey("/kv")))
	return &inmemRepo{
		c:  c,
		ds: ds,
		st: st,
	}, nil
}

func (i *inmemRepo) Config() config.Config {
	return i.c
}

func (i *inmemRepo) SetConfig(c config.Config) error {
	i.c = c
	return nil
}

func (i *inmemRepo) Datastore() datastore.Batching {
	return i.ds
}

func (i *inmemRepo) Store() gkvstore.Store {
	return i.st
}

func (i *inmemRepo) Close() error {
	i.ds.Close()
	i.st.Close()
	return nil
}

func (i *inmemRepo) Status() interface{} {
	return "In-mem repository"
}
