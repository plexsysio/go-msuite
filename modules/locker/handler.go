package locker

import (
	"errors"
	"github.com/aloknerurkar/dLocker"
	inmem "github.com/aloknerurkar/dLocker/handlers/memlock"
	rd "github.com/aloknerurkar/dLocker/handlers/redis"
	zk "github.com/aloknerurkar/dLocker/handlers/zookeeper"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewLocker),
)

func NewLocker(c config.Config) (dLocker.DLocker, error) {
	var lk string
	ok := c.Get("Locker", &lk)
	if !ok {
		return nil, errors.New("Locker not configured")
	}
	switch lk {
	case "inmem":
		return inmem.NewLocker(), nil
	case "zookeeper":
		var host string
		var port int
		if ok := c.Get("ZookeeperHost", &host); !ok {
			return nil, errors.New("Zookeeper host absent")
		}
		if ok := c.Get("ZookeeperPort", &port); !ok {
			return nil, errors.New("Zookeeper port absent")
		}
		return zk.NewZkLocker(host, port)
	case "redis":
		var host, netw string
		if ok := c.Get("RedisHost", &host); !ok {
			return nil, errors.New("Redis host absent")
		}
		if ok := c.Get("RedisNetwork", &netw); !ok {
			return nil, errors.New("Redis host absent")
		}
		// TODO: Add config for username/password authentication
		return rd.NewRedisLocker(netw, host), nil
	}
	return nil, errors.New("Invalid locker handler")
}
