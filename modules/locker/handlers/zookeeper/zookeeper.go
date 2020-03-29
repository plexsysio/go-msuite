//     Digota <http://digota.com> - eCommerce microservice
//     Copyright (C) 2017  Yaron Sumel <yaron@digota.com>. All Rights Reserved.
//
//     This program is free software: you can redistribute it and/or modify
//     it under the terms of the GNU Affero General Public License as published
//     by the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.
//
//     This program is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU Affero General Public License for more details.
//
//     You should have received a copy of the GNU Affero General Public License
//     along with this program.  If not, see <http://www.gnu.org/licenses/>.

// watch dump every second:
// watch -n 1 -d '{ echo "dump"; sleep 1; } | telnet localhost 2181'

package zookeeper

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/aloknerurkar/go-msuite/modules/locker"
	storeItem "github.com/aloknerurkar/go-msuite/modules/store"
	"github.com/aloknerurkar/go-msuite/utils"
	logger "github.com/ipfs/go-log"
	"github.com/yaronsumel/go-zookeeper/zk"
)

var log = logger.Logger("locker/zk")

const separator = "/"

type zkLocker struct {
	zp  *utils.Pool
	mtx sync.Mutex
}

func getZNodePath(obj storeItem.Item) (string, error) {
	if obj.GetNamespace() == "" || obj.GetId() == "" {
		return "", errors.New("Obj is missing information to make that lock")
	}

	path := separator + obj.GetNamespace() + separator + obj.GetId()
	log.Debugf("lock path: %s", path)

	return separator + obj.GetNamespace() + separator + obj.GetId(), nil
}

func NewZkLocker(conf config.Config) (locker.Locker, error) {
	zookeeperHost, ok := conf.Get("zookeeper_hostname").(string)
	if !ok {
		return nil, errors.New("zookeeper hostname missing")
	}
	zookeeperPort, ok := conf.Get("zookeeper_port").(int)
	if !ok {
		return nil, errors.New("zookeeper port missing")
	}
	p := newPool(fmt.Sprintf("%s:%d", zookeeperHost, zookeeperPort))
	return &zkLocker{zp: p}, nil
}

func newPool(serverAddr string) *utils.Pool {
	return &utils.Pool{
		DialFn: func(addr string) (utils.Conn, error) {
			c, _, err := zk.Connect([]string{addr}, time.Millisecond*100)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		HeartBeatFn: func(conn utils.Conn) error {
			if zkConnObj, ok := conn.(*zk.Conn); ok {
				if zkConnObj.State() == zk.StateDisconnected ||
					zkConnObj.State() == zk.StateExpired {
					return errors.New("Connection state wrong")
				}
			}
			return nil
		},
		CloseFn: func(conn utils.Conn) error {
			if zkConnObj, ok := conn.(*zk.Conn); ok {
				zkConnObj.Close()
			}
			return nil
		},
		MaxConnections: 50,
		MaxIdle:        10,
		ServerAddr:     serverAddr,
	}
}

func (l *zkLocker) newLock(obj storeItem.Item) (*zk.Lock, func(), error) {
	znodePath, _ := getZNodePath(obj)
	conn, done, err := l.zp.GetConn()
	if err != nil {
		return nil, nil, err
	}

	if zkConnObj, ok := conn.(*zk.Conn); ok {
		return zk.NewLock(zkConnObj, znodePath, zk.WorldACL(zk.PermAll)), done, nil
	}
	return nil, done, errors.New("Unexpected conn received")
}

func (l *zkLocker) Close() error {
	// All the ZK connections are closed as soon as TryLock is completed.
	// So, just returning nil here
	return nil
}

func (l *zkLocker) Lock(obj storeItem.Item) (func() error, error) {
	z, done, err := l.newLock(obj)
	if err != nil {
		return nil, err
	}
	if err := z.Lock(); err != nil {
		return nil, err
	}
	return func() error {
		defer done()
		return z.Unlock()
	}, nil
}

func (l *zkLocker) TryLock(obj storeItem.Item, t time.Duration) (func() error, error) {
	z, done, err := l.newLock(obj)
	if err != nil {
		return nil, err
	}

	if err := z.TryLock(t); err != nil {
		return nil, err
	}
	return func() error {
		defer done()
		return z.Unlock()
	}, nil
}
