package fsrepo

import (
	"encoding/base64"
	"errors"
	"github.com/StreamSpace/ss-store"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/aloknerurkar/go-msuite/modules/repo"
	"github.com/aloknerurkar/go-msuite/utils"
	ds "github.com/ipfs/go-datastore"
	ci "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"os"
	"sync"
)

var (
	pkgLock sync.Mutex
)

type fsRepo struct {
	path string
}

func datastorePath(root string) string {
	return filepath.Join(root, "datastore")
}

func initRepo(path string) error {
	err := utils.MkdirIfNotExists(path)
	if err != nil {
		return errors.New("fsRepo: failed creating directory Err: " + err.Error())
	}

	return nil
}

func initIdentity(c config.Config) error {
	sk, pk, err := ci.GenerateKeyPair(ci.Ed25519, 2048)
	if err != nil {
		return err
	}
	skbytes, err := sk.Bytes()
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

func Init(path string, c config.Config) error {
	pkgLock.Lock()
	defer pkgLock.Unlock()
	// Check if config is already present
	if utils.Exists(c.FileName(path)) {
		return errors.New("fsRepo: already initialized")
	}
	// Write the initial config provided
	err := utils.WriteToFile(c, c.FileName(path))
	if err != nil {
		return errors.New("fsRepo: failed creating config Err: %s" + err.Error())
	}

	return nil
}

func Open(path string) (repo.Repo, error) {
	return nil, nil
}

func (f *fsRepo) Config() config.Config {
	return nil
}

func (f *fsRepo) SetConfig(c config.Config) error {
	return nil
}

func (f *fsRepo) Store() store.Store {
	return nil
}

func (f *fsRepo) Datastore() ds.Batching {
	return nil
}

func (f *fsRepo) Close() error {
	return nil
}
