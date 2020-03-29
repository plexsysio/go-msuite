package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/aloknerurkar/go-msuite/modules/store"
	"github.com/garyburd/redigo/redis"
	logger "github.com/ipfs/go-log"
	"github.com/satori/go.uuid"
	"strconv"
	"sync"
	"time"
)

var log = logger.Logger("store/redis")

type redisHandler struct {
	addr   string
	dbName string
	pool   *redis.Pool
}

func NewRedisStore(conf config.Config) (store.Store, error) {
	redisHost, ok := conf.Get("redis_hostname").(string)
	if !ok {
		return nil, errors.New("Redis hostname missing")
	}
	redisPort, ok := conf.Get("redis_port").(int)
	if !ok {
		return nil, errors.New("Redis port missing")
	}
	dbName, ok := conf.Get("redis_dbname").(string)
	if !ok {
		dbName = ""
	}

	addr := fmt.Sprintf("%s:%d", redisHost, redisPort)

	store := new(redisHandler)
	store.addr = addr
	store.dbName = dbName
	store.pool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,

		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				return nil, err
			}
			return c, err
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	return store, nil
}

func (r *redisHandler) storeKey(i store.Item) string {
	if len(r.dbName) > 0 {
		return r.dbName + "_" + i.GetNamespace() + "_" + i.GetId()
	}
	return i.GetNamespace() + "_" + i.GetId()
}

func (r *redisHandler) Create(i store.Item) error {
	_conn := r.pool.Get()
	defer _conn.Close()

	if v, ok := i.(store.IdSetter); ok {
		v.SetId(uuid.NewV4().String())
	}

	if v, ok := i.(store.TimeTracker); ok {
		v.SetCreated(time.Now().Unix())
		v.SetUpdated(time.Now().Unix())
		_, err := _conn.Do("ZADD", "created_"+i.GetNamespace(),
			v.GetCreated(), r.storeKey(i))
		if err != nil {
			return fmt.Errorf("error creating created time index %s: %v",
				r.storeKey(i), err)
		}
		_, err = _conn.Do("ZADD", "updated_"+i.GetNamespace(),
			v.GetUpdated(), r.storeKey(i))
		if err != nil {
			return fmt.Errorf("error creating updated time index %s: %v",
				r.storeKey(i), err)
		}
		log.Info("Added created and updated index")
	}

	if msg, ok := i.(store.Exportable); ok {
		val, err := msg.Marshal()
		if err != nil {
			return fmt.Errorf("error marshalling key %s: %v", r.storeKey(i), err)
		}

		_, err = _conn.Do("SET", r.storeKey(i), val)
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

func (r *redisHandler) Update(i store.Item) error {
	_conn := r.pool.Get()
	defer _conn.Close()

	if v, ok := i.(store.TimeTracker); ok {
		v.SetUpdated(time.Now().Unix())
		_, err := _conn.Do("ZREM", "updated_"+i.GetNamespace(), r.storeKey(i))
		if err != nil {
			return fmt.Errorf("error removing updated time index %s: %v",
				r.storeKey(i), err)
		}
		_, err = _conn.Do("ZADD", "updated_"+i.GetNamespace(),
			v.GetUpdated(), r.storeKey(i))
		if err != nil {
			return fmt.Errorf("error updating updated time index %s: %v",
				r.storeKey(i), err)
		}
	}

	if msg, ok := i.(store.Exportable); ok {
		val, err := msg.Marshal()
		if err != nil {
			return fmt.Errorf("error marshalling key %s: %v", r.storeKey(i), err)
		}

		_, err = _conn.Do("SET", r.storeKey(i), val)
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

func (r *redisHandler) Delete(i store.Item) error {
	_conn := r.pool.Get()
	defer _conn.Close()

	_, err := _conn.Do("DEL", r.storeKey(i))
	if err != nil {
		return fmt.Errorf("error deleting key %s: %v", r.storeKey(i), err)
	}
	return err
}

func (r *redisHandler) Read(i store.Item) error {
	_conn := r.pool.Get()
	defer _conn.Close()

	val, err := redis.Bytes(_conn.Do("GET", r.storeKey(i)))
	if err != nil {
		return fmt.Errorf("error getting key %s: %v", r.storeKey(i), err)
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

func getKeys(conn redis.Conn, o store.ListOpt,
	ns string, errc chan error, comm chan string) (count int) {
	count = 0
	skip := o.Page * o.Limit
	handleArr := func(k []string) bool {
		if int64(len(k)) < skip {
			skip -= int64(len(k))
		} else {
			for _, v := range k[skip:] {
				log.Debugf("Sending %s", v)
				comm <- v
				count++
				if int64(count) == o.Limit {
					log.Debugf("Sent %d items on channel", count)
					return true
				}
			}
			skip = 0
		}
		return false
	}
	switch o.Sort {
	case store.SortNatural:
		log.Debugf("Natural sort")
		pattern := ns + "_*"
		iter := 0
		for {
			arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern))
			if err != nil {
				errc <- err
				return
			}
			iter, _ = redis.Int(arr[0], nil)
			keys, _ := redis.Strings(arr[1], nil)
			if handleArr(keys) {
				return
			}
			if iter == 0 {
				return
			}
		}
	case store.SortCreatedAsc:
		fallthrough
	case store.SortCreatedDesc:
		version := "-inf"
		if o.Version != 0 {
			version = strconv.FormatInt(o.Version, 10)
		}
		log.Infof("Using version %s", version)
		arr, err := redis.Strings(conn.Do("ZRANGEBYSCORE",
			"created_"+ns, version, "+inf"))
		if err != nil {
			errc <- err
			return
		}
		_ = handleArr(arr)
	case store.SortUpdatedAsc:
		fallthrough
	case store.SortUpdatedDesc:
		version := "-inf"
		if o.Version != 0 {
			version = strconv.FormatInt(o.Version, 10)
		}
		arr, err := redis.Strings(conn.Do("ZRANGEBYSCORE",
			"updated_"+ns, version, "+inf"))
		if err != nil {
			errc <- err
			return
		}
		_ = handleArr(arr)
	}
	return
}

func (r *redisHandler) List(l store.Items, o store.ListOpt) (int, error) {

	if int64(len(l)) < o.Limit {
		return 0, fmt.Errorf("error insufficient items in array to unmarshal required %d got %d",
			o.Limit, len(l))
	}

	comm := make(chan string, o.Limit)
	errc := make(chan error, 1)

	ctx, cancel := context.WithCancel(context.Background())
	count := 0

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		conn := r.pool.Get()
		defer conn.Close()
		defer wg.Done()
		defer cancel()
		count = getKeys(conn, o, l[0].GetNamespace(), errc, comm)
	}()

	idx := 0
	var retErr error
	wg.Add(1)
	go func() {
		conn := r.pool.Get()
		defer conn.Close()
		defer wg.Done()
		for {
			select {
			case _key := <-comm:
				log.Debugf("Got key %s", _key)
				val, err := redis.Bytes(conn.Do("GET", _key))
				if err == nil {
					if msg, ok := l[idx].(store.Exportable); ok {
						err = msg.Unmarshal(val)
					} else {
						err = fmt.Errorf("unsupported object type: %v", l[idx])
					}
				}
				if err != nil {
					retErr = err
					return
				}
				idx++
			case e := <-errc:
				retErr = e
				return
			case <-ctx.Done():
				if idx == count {
					log.Debugf("Done")
					return
				}
				log.Debugf("Have to drain requests first idx %d count %d", idx, count)
			}
		}
	}()

	wg.Wait()

	return idx, retErr
}
