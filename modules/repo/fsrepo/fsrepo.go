package fsrepo

import (
	"encoding/base64"
	"errors"
	"fmt"
	ssStore "github.com/StreamSpace/ss-ds-store"
	"github.com/StreamSpace/ss-store"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/aloknerurkar/go-msuite/modules/config/json"
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
	active *fsRepo
	refCnt int
}

func (r *repoOpener) Open(path string) (repo.Repo, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.active != nil {
		r.refCnt++
		return r.active, nil
	}
	rp, err := open(path)
	if err != nil {
		return nil, err
	}
	r.active = rp.(*fsRepo)
	r.refCnt = 1
	return r.active, nil
}

func (r *repoOpener) Close() error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.refCnt--
	if r.refCnt > 0 {
		return nil
	}
	return r.active.close()
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
	err := fsRepoErr{msg: msg}
	if secErr != nil {
		err.Append(secErr)
	}
	return err
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
	fmt.Println("ID SET")
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
		return wrapError("already initialized", nil)
	}
	if err := initRepo(path); err != nil {
		fmt.Println(path, err.Error())
		return wrapError("failed creating directories", err)
	}
	// Create new IDs if not provided
	id := map[string]interface{}{}
	if !c.Get("Identity", &id) {
		if err := initIdentity(c); err != nil {
			return wrapError("failed creating identity", err)
		}
	} else {
		fmt.Println("IDENTITY PRESENT")
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
		return nil, wrapError("not initialized", nil)
	}
	return opener.Open(path)
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

func CreateOrOpen(path string, c config.Config) (repo.Repo, error) {
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

	return opener.Close()
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
	f.kvStore = nil
	f.rootDS = nil
	return nil
}
