package boltdb

import (
	"errors"
	"fmt"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/aloknerurkar/go-msuite/modules/store"
	"github.com/boltdb/bolt"
	"github.com/satori/go.uuid"
	"os"
	"sync"
	"time"
)

var DB_FILE = "bolt_store.db"
var BUCKET = "main"

type boltdbHandler struct {
	dbP *bolt.DB
	lk  sync.Mutex
}

func NewBoltDbStore(conf config.Config) (*boltdbHandler, error) {
	dbPath, ok := conf.Get("bolt_path").(string)
	if !ok {
		return nil, errors.New("Bolt path missing")
	}
	if _, e := os.Stat(dbPath); e != nil {
		return nil, e
	}
	fullName := dbPath + "/" + DB_FILE
	db, e := bolt.Open(fullName, 0600, nil)
	if e != nil {
		return nil, e
	}
	store := new(boltdbHandler)
	store.dbP = db

	return store, nil
}

func (r *boltdbHandler) storeKey(i store.Item) string {
	return i.GetNamespace() + "_" + i.GetId()
}

func (r *boltdbHandler) Create(i store.Item) error {
	r.lk.Lock()
	defer r.lk.Unlock()

	if v, ok := i.(store.IdSetter); ok {
		v.SetId(uuid.NewV4().String())
	}

	if v, ok := i.(store.TimeTracker); ok {
		v.SetCreated(time.Now().Unix())
		v.SetUpdated(time.Now().Unix())
	}

	if msg, ok := i.(store.Exportable); ok {
		val, err := msg.Marshal()
		if err != nil {
			return fmt.Errorf("error marshalling key %s: %v", r.storeKey(i), err)
		}
		err = r.dbP.Update(func(tx *bolt.Tx) error {
			bkt, err := tx.CreateBucketIfNotExists([]byte(BUCKET))
			if err != nil {
				return err
			}
			err = bkt.Put([]byte(r.storeKey(i)), val)
			return err
		})
		if err != nil {
			v := string(val)
			if len(v) > 15 {
				v = v[0:12] + "..."
			}
			return fmt.Errorf("error setting key %s to %s: %v", r.storeKey(i), v, err)
		}
		return nil
	}

	return fmt.Errorf("unsupported object type: %v", i)
}

func (r *boltdbHandler) Update(i store.Item) error {
	r.lk.Lock()
	defer r.lk.Unlock()

	if v, ok := i.(store.TimeTracker); ok {
		v.SetUpdated(time.Now().Unix())
	}

	if msg, ok := i.(store.Exportable); ok {
		val, err := msg.Marshal()
		if err != nil {
			return fmt.Errorf("error marshalling key %s: %v", r.storeKey(i), err)
		}
		err = r.dbP.Update(func(tx *bolt.Tx) error {
			bkt := tx.Bucket([]byte(BUCKET))
			if bkt == nil {
				return errors.New("Bucket does not exist")
			}
			err = bkt.Put([]byte(r.storeKey(i)), val)
			return err
		})
		if err != nil {
			v := string(val)
			if len(v) > 15 {
				v = v[0:12] + "..."
			}
			return fmt.Errorf("error setting key %s to %s: %v", r.storeKey(i), v, err)
		}
		return nil
	}

	return fmt.Errorf("unsupported object type: %v", i)
}

func (r *boltdbHandler) Delete(i store.Item) error {
	r.lk.Lock()
	defer r.lk.Unlock()

	return r.dbP.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(BUCKET))
		if bkt == nil {
			return errors.New("Bucket does not exist")
		}
		err := bkt.Delete([]byte(r.storeKey(i)))
		return err
	})
}

func (r *boltdbHandler) Read(i store.Item) error {
	r.lk.Lock()
	defer r.lk.Unlock()

	var val []byte
	err := r.dbP.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(BUCKET))
		if bkt == nil {
			return bolt.ErrBucketNotFound
		}
		val = bkt.Get([]byte(r.storeKey(i)))
		if val == nil {
			return fmt.Errorf("error getting key %s", r.storeKey(i))
		}
		return nil
	})
	if err != nil {
		return err
	}
	if msg, ok := i.(store.Exportable); ok {
		err = msg.Unmarshal(val)
		if err != nil {
			v := string(val)
			if len(v) > 15 {
				v = v[0:12] + "..."
			}
			return fmt.Errorf("error unmarshalling data %s to %s: %v", v, r.storeKey(i), err)
		}
		return nil
	}

	return fmt.Errorf("unsupported object type: %v", i)
}

func (r *boltdbHandler) List(l store.Items, o store.ListOpt) (int, error) {
	r.lk.Lock()
	defer r.lk.Unlock()

	if int64(len(l)) < o.Limit {
		return 0, fmt.Errorf("error insufficient items in array to unmarshal required %d got %d",
			o.Limit, len(l))
	}
	idx := 0
	err := r.dbP.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(BUCKET))
		if bkt == nil {
			return bolt.ErrBucketNotFound
		}

		c := bkt.Cursor()
		skip := o.Page * o.Limit

		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			if skip > 0 {
				skip--
				continue
			}
			if msg, ok := l[idx].(store.Exportable); !ok {
				return fmt.Errorf("unsupported object type: %v", l[idx])
			} else {
				err := msg.Unmarshal(v)
				if err != nil {
					return err
				}
			}
			idx++
		}
		return nil
	})

	return idx, err
}
