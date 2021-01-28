package fsrepo

import (
	"encoding/base64"
	"errors"
	"fmt"
	ssStore "github.com/StreamSpace/ss-ds-store"
	"github.com/StreamSpace/ss-store"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/aloknerurkar/go-msuite/modules/repo"
	"github.com/aloknerurkar/go-msuite/utils"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	ci "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"path/filepath"
	"sync"
)

type repoOpener struct {
	mtx    sync.Mutex
	active repo.Repo
	refCnt int
}

func (r *repoOpener) Open(openFn func() (repo.Repo, error)) (repo.Repo, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	defer func() {
		r.refCnt++
	}()
	if r.active != nil {
		return r.active, nil
	}
	rp, err := openFn()
	if err != nil {
		return nil, err
	}
	r.active = rp
	r.refCnt = 0
	return r.active, nil
}

func (r *repoOpener) Close(closeFn func() error) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.refCnt--
	if r.refCnt > 0 {
		return nil
	}
	return closeFn()
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

type fsRepoErr struct {
	msg    string
	secErr []error
}

func (f fsRepoErr) Error() string {
	if f.secErr != nil && len(f.secErr) > 0 {
		errStr := ""
		for i, v := range f.secErr {
			errStr += fmt.Sprintf("[%d] %s\t", i, v.Error())
		}
		return fmt.Sprintf("fsRepo: %s SecErr: %s", f.msg, errStr)
	}
	return fmt.Sprintf("fsRepo: %s", f.msg)
}

func (f fsRepoErr) Append(err error) {
	f.secErr = append(f.secErr, err)
}

func (f fsRepoErr) HasSecErr() bool {
	return f.secErr != nil && len(f.secErr) > 0
}

func wrapError(msg string, secErr error) fsRepoErr {
	return fsRepoErr{msg: msg, secErr: []error{secErr}}
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

func isInitialized(path string) bool {
	return utils.Exists(configPath(path))
}

func Init(path string, c config.Config) error {
	pkgLock.Lock()
	defer pkgLock.Unlock()
	// Check if config is already present
	if isInitialized(path) {
		return wrapError("already initialized", nil)
	}
	if err := initRepo(path); err != nil {
		return wrapError("failed creating directories", err)
	}
	// Create new IDs if not provided
	if !c.Get("Identity", &map[string]interface{}{}) {
		if err := initIdentity(c); err != nil {
			return wrapError("failed creating identity", err)
		}
	}
	// Write the initial config provided
	err := utils.WriteToFile(c, configPath(path))
	if err != nil {
		return wrapError("failed creating config", err)
	}
	return nil
}

func (f *fsRepo) openConfig() error {
	if !utils.Exists(configPath(f.path)) {
		return errors.New("config is absent")
	}
	cfg, err := config.FromFile(configPath(f.path))
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
		return nil, wrapError("not initialized", nil)
	}
	return opener.Open(func() (repo.Repo, error) {
		return open(path)
	})
}

func open(path string) (repo.Repo, error) {
	r := &fsRepo{
		path: path,
	}
	if err := r.openConfig(); err != nil {
		return nil, wrapError("failed opening config", err)
	}
	if err := r.openDatastore(); err != nil {
		return nil, wrapError("failed opening datastore", err)
	}
	if err := r.openStore(); err != nil {
		return nil, wrapError("failed opening KV store", err)
	}
	return r, nil
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
	return utils.WriteToFile(c, configPath(f.path))
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

	return opener.Close(func() error {
		return f.close()
	})
}

func (f *fsRepo) close() error {
	err := wrapError("failed closing repo", nil)
	if f.kvStore != nil {
		e := f.kvStore.Close()
		if e != nil {
			err.Append(e)
		}
	}
	if f.rootDS != nil {
		e := f.rootDS.Close()
		if e != nil {
			err.Append(e)
		}
	}
	if err.HasSecErr() {
		return err
	}
	return nil
}
