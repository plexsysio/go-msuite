package fsrepo

import (
	"encoding/base64"
	"errors"
	"fmt"
	ssStore "github.com/SWRMLabs/ss-ds-store"
	"github.com/SWRMLabs/ss-store"
	"github.com/hashicorp/go-multierror"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	ci "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/plexsysio/go-msuite/modules/config"
	"github.com/plexsysio/go-msuite/modules/config/json"
	"github.com/plexsysio/go-msuite/modules/repo"
	"github.com/plexsysio/go-msuite/utils"
	"path/filepath"
	"sync"
)

type repoOpener struct {
	mtx    sync.Mutex
	Active *fsRepo
	RefCnt int
}

func (r *repoOpener) Open(path string) (repo.Repo, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.Active != nil {
		r.RefCnt++
		return r.Active, nil
	}
	rp, err := open(path)
	if err != nil {
		return nil, err
	}
	r.Active = rp.(*fsRepo)
	r.RefCnt = 1
	return r.Active, nil
}

func (r *repoOpener) Close() error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.RefCnt--
	if r.RefCnt > 0 {
		return nil
	}
	ar := r.Active
	r.Active = nil
	return ar.close()
}

var (
	pkgLock     sync.Mutex
	opener      = &repoOpener{}
	storePrefix = ds.NewKey("s")
)

type fsRepo struct {
	path    string
	cfg     config.Config
	rootDS  ds.Batching
	kvStore store.Store
}

func datastorePath(root string) string {
	return filepath.Join(root, "datastore")
}

func configPath(root string) string {
	return filepath.Join(root, "config.json")
}

func initRepo(path string) error {
	err := utils.MkdirIfNotExists(path)
	if err != nil {
		return err
	}
	err = utils.MkdirIfNotExists(datastorePath(path))
	if err != nil {
		return err
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

func IsInitialized(path string) bool {
	pkgLock.Lock()
	defer pkgLock.Unlock()
	return isInitialized(path)
}

func isInitialized(path string) bool {
	return utils.Exists(configPath(path))
}

func Init(path string, c config.Config) error {
	pkgLock.Lock()
	defer pkgLock.Unlock()
	// Check if config is already present
	if isInitialized(path) {
		return errors.New("already initialized")
	}
	if err := initRepo(path); err != nil {
		return fmt.Errorf("failed creating directories %w", err)
	}
	// Create new IDs if not provided
	id := map[string]interface{}{}
	if !c.Get("Identity", &id) {
		if err := initIdentity(c); err != nil {
			return fmt.Errorf("failed creating identity %w", err)
		}
	}
	// Write the initial config provided
	confRdr, err := c.Reader()
	if err != nil {
		return fmt.Errorf("failed reading config %w", err)
	}
	err = utils.WriteToFile(confRdr, configPath(path))
	if err != nil {
		return fmt.Errorf("failed creating config %w", err)
	}
	return nil
}

func (f *fsRepo) openConfig() error {
	if !utils.Exists(configPath(f.path)) {
		return errors.New("config is absent")
	}
	cfg, err := jsonConf.FromFile(configPath(f.path))
	if err != nil {
		return err
	}
	f.cfg = cfg
	return nil
}

func (f *fsRepo) openDatastore() error {
	if !utils.Exists(datastorePath(f.path)) {
		return utils.MkdirIfNotExists(datastorePath(f.path))
	}
	ds, err := openDatastoreFromCfg(f.path, f.cfg)
	if err != nil {
		return err
	}
	f.rootDS = ds
	return nil
}

func (f *fsRepo) openStore() error {
	nds := namespace.Wrap(f.rootDS, storePrefix)
	dsStore, err := ssStore.NewDataStore(&ssStore.DSConfig{
		DS: nds,
	})
	if err != nil {
		return err
	}
	f.kvStore = dsStore
	return nil
}

func Open(path string) (repo.Repo, error) {
	pkgLock.Lock()
	defer pkgLock.Unlock()

	if !isInitialized(path) {
		return nil, errors.New("not initialized")
	}
	return opener.Open(path)
}

func open(path string) (repo.Repo, error) {
	r := &fsRepo{
		path: path,
	}
	if err := r.openConfig(); err != nil {
		return nil, fmt.Errorf("failed opening config %w", err)
	}
	if err := r.openDatastore(); err != nil {
		return nil, fmt.Errorf("failed opening datastore %w", err)
	}
	if err := r.openStore(); err != nil {
		return nil, fmt.Errorf("failed opening KV store %w", err)
	}
	return r, nil
}

func CreateOrOpen(c config.Config) (repo.Repo, error) {
	var path string
	found := c.Get("RootPath", &path)
	if !found {
		return nil, errors.New("root path not specified")
	}
	if !IsInitialized(path) {
		err := Init(path, c)
		if err != nil {
			return nil, err
		}
	}
	return Open(path)
}

func (f *fsRepo) Config() config.Config {
	pkgLock.Lock()
	defer pkgLock.Unlock()
	return f.cfg
}

func (f *fsRepo) SetConfig(c config.Config) error {
	pkgLock.Lock()
	defer pkgLock.Unlock()
	f.cfg = c
	confRdr, err := c.Reader()
	if err != nil {
		return err
	}
	return utils.WriteToFile(confRdr, configPath(f.path))
}

func (f *fsRepo) Store() store.Store {
	pkgLock.Lock()
	defer pkgLock.Unlock()
	return f.kvStore
}

func (f *fsRepo) Datastore() ds.Batching {
	pkgLock.Lock()
	defer pkgLock.Unlock()
	return f.rootDS
}

func (f *fsRepo) Close() error {
	pkgLock.Lock()
	defer pkgLock.Unlock()

	return opener.Close()
}

func (f *fsRepo) close() error {
	var err *multierror.Error
	if f.kvStore != nil {
		e := f.kvStore.Close()
		if e != nil {
			err = multierror.Append(err, e)
		}
	}
	if f.rootDS != nil {
		e := f.rootDS.Close()
		if e != nil {
			err = multierror.Append(err, e)
		}
	}
	f.kvStore = nil
	f.rootDS = nil
	return err.ErrorOrNil()
}
