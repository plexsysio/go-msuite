package grpcmux

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	logger "github.com/ipfs/go-log/v2"
	"github.com/plexsysio/taskmanager"
)

var log = logger.Logger("grpc/lmux")

type MuxListener struct {
	Tag      string
	Listener net.Listener
}

type Mux struct {
	muxCtx    context.Context
	muxCancel context.CancelFunc
	listeners []MuxListener
	tm        *taskmanager.TaskManager
	connChan  chan net.Conn
	wg        sync.WaitGroup
	statusMtx sync.Mutex
	status    map[string]interface{}
}

func New(
	ctx context.Context,
	listeners []MuxListener,
	tm *taskmanager.TaskManager,
) *Mux {
	muxCtx, muxCancel := context.WithCancel(ctx)
	m := &Mux{
		muxCtx:    muxCtx,
		muxCancel: muxCancel,
		listeners: listeners,
		tm:        tm,
		connChan:  make(chan net.Conn, 50),
		status:    make(map[string]interface{}),
	}
	for _, v := range listeners {
		m.updateStatus(v.Tag, "not running")
	}
	return m
}

func (m *Mux) updateStatus(key, value string) {
	m.statusMtx.Lock()
	defer m.statusMtx.Unlock()

	m.status[key] = value
}

func (m *Mux) Status() interface{} {
	m.statusMtx.Lock()
	defer m.statusMtx.Unlock()

	return m.status
}

func (m *Mux) Start(ctx context.Context) {
	for i := range m.listeners {
		l := &muxListener{
			tag:      m.listeners[i].Tag,
			listener: m.listeners[i].Listener,
			connChan: m.connChan,
			reportErr: func(k string, err error) {
				m.updateStatus(k, "failed with err: "+err.Error())
			},
		}
		m.wg.Add(1)
		sched, err := m.tm.GoFunc(l.Name(), func(c context.Context) error {
			defer m.wg.Done()
			return l.Execute(c)
		})
		if err != nil {
			m.updateStatus(m.listeners[i].Tag, "failed to start err: "+err.Error())
			continue
		}
		select {
		case <-sched:
			m.updateStatus(m.listeners[i].Tag, "running")
		case <-ctx.Done():
			return
		}
	}
}

func (m *Mux) Accept() (net.Conn, error) {
	select {
	case <-m.muxCtx.Done():
		log.Info("Context cancelled")
		return nil, io.EOF
	case newConn := <-m.connChan:
		log.Info("Handling new connection")
		return newConn, nil
	}
}

func (m *Mux) Close() error {
	stopped := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(stopped)
	}()

	m.muxCancel()
	var err *multierror.Error
	for _, l := range m.listeners {
		e := l.Listener.Close()
		if e != nil {
			err = multierror.Append(err, e)
		}
	}

	select {
	case <-stopped:
	case <-time.After(3 * time.Second):
		err = multierror.Append(err, errors.New("failed to stop listeners"))
	}

	return err.ErrorOrNil()
}

func (m *Mux) Addr() net.Addr {
	if len(m.listeners) > 0 {
		return m.listeners[0].Listener.Addr()
	}
	return &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: 0}
}

type muxListener struct {
	tag       string
	listener  net.Listener
	connChan  chan<- net.Conn
	reportErr func(string, error)
}

func (m *muxListener) Name() string {
	return "MuxListener_" + m.tag
}

func (m *muxListener) Execute(ctx context.Context) error {
	for {
		conn, err := m.listener.Accept()
		if err != nil {
			log.Error("Failed accepting new connection from listener", m.tag, err.Error())
			m.reportErr(m.tag, err)
			return err
		}
		select {
		case <-ctx.Done():
			log.Info("Closing mux listener", m.tag)
			return nil
		case m.connChan <- conn:
			log.Info("Enqueued connection from listener", m.tag)
		}
	}
}
