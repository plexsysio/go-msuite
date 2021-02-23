package locker

import (
	"context"
	"errors"
	"github.com/aloknerurkar/dLocker"
	inmem "github.com/aloknerurkar/dLocker/handlers/memlock"
	rd "github.com/aloknerurkar/dLocker/handlers/redis"
	zk "github.com/aloknerurkar/dLocker/handlers/zookeeper"
	"github.com/aloknerurkar/go-msuite/modules/config"
	logger "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"
)

var log = logger.Logger("locker")

var Module = fx.Options(
	fx.Provide(NewLocker),
)

func NewLocker(
	lc fx.Lifecycle,
	c config.Config,
) (dLocker.DLocker, error) {
	var lk string
	ok := c.Get("Locker", &lk)
	if !ok {
		return nil, errors.New("Locker not configured")
	}
	var lkr dLocker.DLocker
	var retErr error
	switch lk {
	case "inmem":
		lkr, retErr = inmem.NewLocker(), nil
	case "zookeeper":
		var host string
		var port int
		if ok := c.Get("ZookeeperHost", &host); !ok {
			return nil, errors.New("Zookeeper host absent")
		}
		if ok := c.Get("ZookeeperPort", &port); !ok {
			return nil, errors.New("Zookeeper port absent")
		}
		lkr, retErr = zk.NewZkLocker(host, port)
	case "redis":
		var host, netw string
		if ok := c.Get("RedisHost", &host); !ok {
			return nil, errors.New("Redis host absent")
		}
		if ok := c.Get("RedisNetwork", &netw); !ok {
			return nil, errors.New("Redis host absent")
		}
		// TODO: Add config for username/password authentication
		lkr, retErr = rd.NewRedisLocker(netw, host), nil
	default:
		return nil, errors.New("Invalid locker handler")
	}
	if retErr == nil {
		log.Info("Configured DLocker handler %s", lk)
		lc.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				defer log.Info("Closed DLocker")
				return lkr.Close()
			},
		})
	}
	return lkr, retErr
}
