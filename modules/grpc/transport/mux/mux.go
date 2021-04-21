package grpcmux

import (
	"context"
	"github.com/SWRMLabs/ss-taskmanager"
	logger "github.com/ipfs/go-log/v2"
	"github.com/plexsysio/go-msuite/modules/diag/status"
	"go.uber.org/fx"
	"io"
	"net"
)

var log = logger.Logger("grpc/lmux")

var Module = fx.Provide(NewMuxedListener)

type MuxListenerOut struct {
	fx.Out

	Listener MuxListener `group:"listener"`
}

type MuxIn struct {
	fx.In

	Listeners []MuxListener  `group:"listener"`
	StManager status.Manager `optional:"true"`
}

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

func NewMuxedListener(
	lc fx.Lifecycle,
	ctx context.Context,
	listeners MuxIn,
	tm *taskmanager.TaskManager,
) (*Mux, error) {
	m := &Mux{
		listeners: listeners.Listeners,
		tm:        tm,
		connChan:  make(chan net.Conn, 50),
	}
	m.muxCtx, m.muxCancel = context.WithCancel(ctx)
	m.start(func(key string, err error) {
		dMap := map[string]interface{}{
			key: "Failed Err:" + err.Error(),
		}
		if listeners.StManager != nil {
			listeners.StManager.Report("RPC Listeners", status.Map(dMap))
		}
	})
	stMp := make(map[string]interface{})
	for _, v := range listeners.Listeners {
		stMp[v.Tag] = "Running"
	}
	if listeners.StManager != nil {
		listeners.StManager.Report("RPC Listeners", status.Map(stMp))
	}
	lc.Append(fx.Hook{
		OnStop: func(c context.Context) error {
			defer func() {
				for k, _ := range stMp {
					stMp[k] = "Stopped"
				}
				if listeners.StManager != nil {
					listeners.StManager.Report("RPC Listeners", status.Map(stMp))
				}
			}()
			log.Info("Stopping Muxed listeners")
			err := m.Close()
			if err != nil {
				log.Warn("Error on closing listeners", err.Error())
			}
			return nil
		},
	})
	return m, nil
}

func (m *Mux) start(reportError func(string, error)) {
	for _, v := range m.listeners {
		l := &muxListener{
			tag:      v.Tag,
			listener: v.Listener,
			connChan: m.connChan,
			reportErr: func(err error) {
				reportError(v.Tag, err)
			},
		}
		m.tm.GoWork(l)
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
