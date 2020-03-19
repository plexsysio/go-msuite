package net

import (
	"errors"
	"fmt"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/libp2p/go-libp2p-core/host"
	"go.uber.org/fx"
	"net"
	"time"
)

var P2P = fx.Option(
	fx.Provide(NewP2PListener),
)

func NewP2PListener(conf config.Config) (net.Listener, error) {
	portVal, ok := conf.Get("grpc_port").(int32)
	if !ok {
		return nil, errors.New("Port absent")
	}

	log.Infof("Starting TCP listener on port %d", portVal)
	listnr, err := net.Listen("tcp", fmt.Sprintf(":%d", portVal))
	if err != nil {
		return nil, err
	}

	return listnr, nil
}

type p2pListener struct {
	host.Host
}

func (p *p2pListener) Accept() (net.Conn, error) {
	return nil, nil
}

func (p *p2pListener) Close() error {
	return nil
}

func (p *p2pListener) Addr() net.Addr {
	return nil
}

type p2pConn struct {
}

func (c *p2pConn) Read(b []byte) (n int, err error) {
	return
}

func (c *p2pConn) Write(b []byte) (n int, err error) {
	return
}

func (c *p2pConn) Close() error {
	return nil
}

func (c *p2pConn) LocalAddr() net.Addr {
	return nil
}

func (c *p2pConn) RemoteAddr() net.Addr {
	return nil
}

func (c *p2pConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *p2pConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *p2pConn) SetWriteDeadline(t time.Time) error {
	return nil
}
