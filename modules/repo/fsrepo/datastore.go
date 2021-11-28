package fsrepo

import (
	"context"
	"errors"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/mount"
	"github.com/plexsysio/go-msuite/modules/config"
	"github.com/plexsysio/go-msuite/utils"

	// "github.com/ipfs/go-ds-badger2"
	"path/filepath"

	flatfs "github.com/ipfs/go-ds-flatfs"
	leveldb "github.com/ipfs/go-ds-leveldb"
)

type MountInfo struct {
	Path   string
	Prefix string
	Usage  uint64
}

type Datastore interface {
	ds.Batching
	Mounts() ([]MountInfo, error)
}

var DefaultMountInfo = map[string]interface{}{
	"level": map[string]interface{}{
		"path":   "kv",
		"prefix": "/",
	},
	"flatfs": map[string]interface{}{
		"path":      "blocks",
		"prefix":    "blocks",
		"shardFunc": "/repo/flatfs/shard/v1/next-to-last/2",
		"sync":      true,
	},
}

func openDatastoreFromCfg(root string, c config.Config) (mDS ds.Batching, retErr error) {
	mntInfo := map[string]interface{}{}
	if ok := c.Get("Mounts", &mntInfo); !ok {
		mntInfo = DefaultMountInfo
	}
	mnts := []mount.Mount{}
	mntInfos := map[string]mount.Mount{}
	defer func() {
		if retErr != nil {
			for _, v := range mnts {
				v.Datastore.Close()
			}
		}
	}()
	for k, v := range mntInfo {
		dCfg, ok := v.(map[string]interface{})
		if !ok {
			retErr = errors.New("Sub DS config missing for datastore")
			return
		}
		prefix, ok := dCfg["prefix"].(string)
		if !ok {
			retErr = errors.New("Prefix missing for datastore")
			return
		}
		path, ok := dCfg["path"].(string)
		if !ok {
			path = root
		} else {
			path = filepath.Join(root, path)
			err := utils.MkdirIfNotExists(path)
			if err != nil {
				retErr = err
				return
			}
		}
		var newDs ds.Batching
		var err error
		switch k {
		case "level":
			newDs, err = leveldb.NewDatastore(path, &leveldb.Options{})
		// case "badger":
		// 	newDs, err = badger.NewDatastore(path, &badger.DefaultOptions)
		case "flatfs":
			sFn, ok := dCfg["shardFunc"].(string)
			if !ok {
				sFn = "/repo/flatfs/shard/v1/next-to-last/2"
			}
			sn, ok := dCfg["sync"].(bool)
			if !ok {
				sn = true
			}
			sf, e := flatfs.ParseShardFunc(sFn)
			if e != nil {
				retErr = e
				return
			}
			newDs, err = flatfs.CreateOrOpen(path, sf, sn)
		default:
			retErr = errors.New("Invalid datastore type")
			return
		}
		if err != nil {
			retErr = err
			return
		}
		newMnt := mount.Mount{
			Prefix:    ds.NewKey(prefix),
			Datastore: newDs,
		}
		mnts = append(mnts, newMnt)
		mntInfos[path] = newMnt
	}
	mDS = mount.New(mnts)
	return &mountedDS{
		Batching: mDS,
		mnts:     mntInfos,
	}, nil
}

type mountedDS struct {
	ds.Batching
	mnts map[string]mount.Mount
}

func (m *mountedDS) Mounts() ([]MountInfo, error) {
	mntInfos := []MountInfo{}
	for k, v := range m.mnts {
		if pds, ok := v.Datastore.(ds.PersistentDatastore); ok {
			usg, err := pds.DiskUsage(context.TODO())
			if err != nil {
				return nil, err
			}
			mntInfos = append(mntInfos, MountInfo{
				Path:   k,
				Prefix: v.Prefix.String(),
				Usage:  usg,
			})
		}
	}
	return mntInfos, nil
}
