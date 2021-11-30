package grpcmux

import (
	"context"
	"io"
	"net"

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
	}
	return m
}

func (m *Mux) Start(ctx context.Context, reportError func(string, error)) {
	for _, v := range m.listeners {
		l := &muxListener{
			tag:      v.Tag,
			listener: v.Listener,
			connChan: m.connChan,
			reportErr: func(err error) {
				if reportError != nil {
					reportError(v.Tag, err)
				}
			},
		}
		sched, err := m.tm.Go(l)
		if err != nil {
			reportError(v.Tag, err)
		}
		select {
		case <-sched:
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
	log.Info("Closing listeners")
	errs := []error{}
	for _, l := range m.listeners {
		err := l.Listener.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	log.Info("Closing Mux")
	m.muxCancel()
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
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
	reportErr func(error)
}

func (m *muxListener) Name() string {
	return "MuxListener_" + m.tag
}

func (m *muxListener) Execute(ctx context.Context) error {
	for {
		conn, err := m.listener.Accept()
		if err != nil {
			log.Error("Failed accepting new connection from listener", m.tag, err.Error())
			m.reportErr(err)
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
