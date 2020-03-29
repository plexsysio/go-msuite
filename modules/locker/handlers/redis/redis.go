package redis

import (
	"errors"
	"fmt"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/aloknerurkar/go-msuite/modules/locker"
	storeItem "github.com/aloknerurkar/go-msuite/modules/store"
	"github.com/garyburd/redigo/redis"
	"sync"
	"time"
)

type redisLocker struct {
	rp  Pool
	mtx sync.Mutex
}

// Pool is an interface over the redis.Pool struct
// to make mockable
type Pool interface {
	Get() redis.Conn
	Close() error
}

const separator = "."

var (
	// ErrTimeout returns when you couldn't make a TryLock call
	ErrTimeout = errors.New("Timeout reached")

	// ErrMissingInfo returns when you have and empty Namespace or Object ID
	ErrMissingInfo = errors.New("Obj is missing information to make that lock")
)

// NewLocker return new redis based lock
func NewRedisLocker(conf config.Config) (locker.Locker, error) {
	redisHost, ok := conf.Get("redis_hostname").(string)
	if !ok {
		return nil, errors.New("Redis hostname missing")
	}
	redisPort, ok := conf.Get("redis_port").(int)
	if !ok {
		return nil, errors.New("Redis port missing")
	}
	p := newPool(fmt.Sprintf("%s:%d", redisHost, redisPort), "")
	return &redisLocker{rp: p}, nil
}

func newPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     10,
		MaxActive:   20,
		IdleTimeout: 240 * time.Second,
		Wait:        true, // Wait for the connection pool, no connection pool exhausted error
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server, redis.DialConnectTimeout(1000*time.Millisecond),
				redis.DialReadTimeout(2000*time.Millisecond),
				redis.DialWriteTimeout(2000*time.Millisecond))
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func getKey(doc storeItem.Item) (string, error) {
	if doc.GetNamespace() == "" || doc.GetId() == "" {
		return "", ErrMissingInfo
	}
	return doc.GetNamespace() + separator + doc.GetId(), nil
}

func (l *redisLocker) Close() error {
	return l.rp.Close()
}

func (l *redisLocker) Lock(doc storeItem.Item) (func() error, error) {
	key, err := getKey(doc)
	if err != nil {
		return nil, err
	}

	conn := l.rp.Get()
	_, err = redis.String(conn.Do("SET", key, "NX"))
	conn.Close()
	if err != nil {
		return nil, err
	}

	return func() error { return l.unlock(key) }, nil
}

func (l *redisLocker) TryLock(doc storeItem.Item, t time.Duration) (func() error, error) {
	key, err := getKey(doc)
	if err != nil {
		return nil, err
	}

	ch := make(chan error)
	l.mtx.Lock()
	conn := l.rp.Get()
	l.mtx.Unlock()
	defer conn.Close()

	go func(c redis.Conn) {
		_, err = redis.String(c.Do("SET", key, "NX"))
		select {
		case ch <- err:
		default:
		}
	}(conn)

	select {
	case err = <-ch:
		if err != nil {
			return nil, err
		}
		return func() error { return l.unlock(key) }, nil
	case <-time.After(t):
		return nil, ErrTimeout
	}
}

func (l *redisLocker) unlock(key string) error {
	conn := l.rp.Get()
	_, err := conn.Do("DEL", key)
	conn.Close()
	return err
}
