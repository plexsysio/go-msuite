package utils

import (
	"errors"
	logger "github.com/ipfs/go-log"
	"sync"
	"sync/atomic"
	"time"
)

var log = logger.Logger("connPool")

const getConnTimeout time.Duration = time.Second * 1
const heartbeatTimeout time.Duration = time.Minute * 15

type Conn interface{}

type Pool struct {
	DialFn         func(addr string) (Conn, error)
	HeartBeatFn    func(Conn) error
	CloseFn        func(Conn) error
	MaxConnections int
	MaxIdle        int
	ServerAddr     string
	mtx            sync.Mutex
	initDone       bool
	heartbeatMap   map[Conn]time.Time
	connChan       chan Conn
	opened         int
	idle           int32
}

type Done func()

func (p *Pool) GetConn() (Conn, Done, error) {

	/* Critical section. Only one routine should initialize at a time.
	 * Getting connection from the channel is a blocking call. So we should
	 * not hold the lock before doing that.
	 */
	p.mtx.Lock()
	if !p.initDone {
		p.connChan = make(chan Conn, p.MaxConnections)
		p.heartbeatMap = make(map[Conn]time.Time, p.MaxConnections)
		// Do lazy init. Only open single connection
		conn, err := p.DialFn(p.ServerAddr)
		if err != nil {
			p.mtx.Unlock()
			log.Errorf("Failed dialing server Err:%s", err)
			return nil, nil, err
		}
		p.connChan <- conn
		p.heartbeatMap[conn] = time.Now()
		p.opened++
		p.idle++
		p.initDone = true
		log.Infof("Done initializing pool. Opened first connection")
	}
	p.mtx.Unlock()

	// Blocking call. Timeout if takes too long. This should indicate we need more connections
	// in the config.
	var newConn Conn
	var err error
	select {
	case newConn = <-p.connChan:
		idl := atomic.AddInt32(&p.idle, -1)
		log.Infof("No of idle connections: %d", idl)
		break
	case <-time.After(getConnTimeout):
		log.Warning("Timed out getting new conn.")
		p.mtx.Lock()
		if p.opened < p.MaxConnections {
			newConn, err = p.DialFn(p.ServerAddr)
			if err != nil {
				p.mtx.Unlock()
				return nil, nil, errors.New("Failed to open new connection")
			}
			p.opened++
			p.heartbeatMap[newConn] = time.Now()
			p.mtx.Unlock()
			log.Debugf("Opened new connection after timeout")
		} else {
			p.mtx.Unlock()
			log.Errorf("Already reached max connections")
			return nil, nil, errors.New("Timeout while getting new connection.")
		}
	}

	if time.Since(p.heartbeatMap[newConn]) > heartbeatTimeout {
		err := p.HeartBeatFn(newConn)
		if err != nil {
			// Get the lock before updating the map.
			p.mtx.Lock()
			delete(p.heartbeatMap, newConn)
			p.mtx.Unlock()
			log.Warningf("Heartbeat failed Err:%s", err.Error())
			newConn, err = p.DialFn(p.ServerAddr)
			if err != nil {
				log.Errorf("Failed to create new conn Err:%s", err.Error())
				return nil, nil, err
			}
		}
		// Get the lock before updating the map.
		p.mtx.Lock()
		p.heartbeatMap[newConn] = time.Now()
		p.mtx.Unlock()
	}
	return newConn, func() { p.connDone(newConn) }, nil
}

func (p *Pool) connDone(conn Conn) {
	if int(p.idle) == p.MaxIdle {
		p.CloseFn(conn)
		log.Infof("Closing connection as passed max idle limit")
		return
	}
	idl := atomic.AddInt32(&p.idle, 1)
	log.Infof("No of idle connections: %d", idl)
	p.connChan <- conn
}
