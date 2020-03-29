package os_filestore

import (
	"errors"
	"fmt"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/aloknerurkar/go-msuite/modules/store"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"os"
)

type osFileStore struct {
	root string
}

func NewOSFileStore(conf config.Config) (store.Store, error) {
	root, ok := conf.Get("filestore_root").(string)
	if !ok {
		return nil, errors.New("Redis hostname missing")
	}
	store := new(osFileStore)
	store.root = root
	err := os.Mkdir(store.root, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	return store, nil
}

func (s *osFileStore) getFullPath(i store.Item) string {
	return s.root + "/" + i.GetNamespace() + "/" + i.GetId()
}

func (s *osFileStore) getParentPath(i store.Item) string {
	return s.root + "/" + i.GetNamespace()
}

func (s *osFileStore) Create(i store.Item) error {

	if v, ok := i.(store.IdSetter); ok {
		v.SetId(uuid.NewV4().String())
	}

	if fi, ok := i.(store.FileItemSetter); ok {
		if _, err := os.Stat(s.getParentPath(i)); os.IsNotExist(err) {
			os.Mkdir(s.getParentPath(i), os.ModePerm)
		}
		fp, err := os.Create(s.getFullPath(i))
		if err != nil {
			return fmt.Errorf("failed creating file %v", err)
		}
		fi.SetFp(fp)
		return nil
	}

	return errors.New("Invalid item type")
}

func (s *osFileStore) Update(i store.Item) error {
	if fi, ok := i.(store.FileItemSetter); ok {
		fp, err := os.OpenFile(s.getFullPath(i), os.O_WRONLY, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed opening file for edit %v", err)
		}
		fi.SetFp(fp)
		return nil
	}

	return errors.New("Invalid item type")
}

func (s *osFileStore) Delete(i store.Item) error {
	return os.Remove(s.getFullPath(i))
}

func (s *osFileStore) Read(i store.Item) error {
	if fi, ok := i.(store.FileItemSetter); ok {
		fp, err := os.Open(s.getFullPath(i))
		if err != nil {
			return fmt.Errorf("failed opening file for read %v", err)
		}
		fi.SetFp(fp)
		return nil
	}

	return errors.New("Invalid item type")
}

func (s *osFileStore) List(l store.Items, o store.ListOpt) (int, error) {
	files, err := ioutil.ReadDir(s.root)
	if err != nil {
		return 0, fmt.Errorf("error opening root dir %s Err:%v", s.root, err)
	}

	if o.Page*o.Limit > int64(len(files)) {
		return 0, errors.New("No more files")
	}

	skip := o.Page * o.Limit
	for i := skip; i < skip+o.Limit+1; i++ {
		var fp *os.File
		if fi, ok := l[i].(store.FileItemSetter); ok {
			fp, err = os.Open(files[i].Name())
			if err == nil {
				fi.SetFp(fp)
			}
		} else {
			err = errors.New("File not provided.")
		}
		if err != nil {
			err = fmt.Errorf("Failed iterating over files Err:%v", err)
			break
		}
	}
	return len(l), err
}
