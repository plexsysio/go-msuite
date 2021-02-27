package p2pgrpc

import (
	inet "github.com/libp2p/go-libp2p-core/network"
	manet "github.com/multiformats/go-multiaddr-net"
	"net"
)

type p2pConn struct {
	inet.Stream
}

func (c *p2pConn) LocalAddr() net.Addr {
	addr, err := manet.ToNetAddr(c.Stream.Conn().LocalMultiaddr())
	if err != nil {
		return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0}
	}
	return addr
}

// RemoteAddr returns the remote address.
func (c *p2pConn) RemoteAddr() net.Addr {
	addr, err := manet.ToNetAddr(c.Stream.Conn().RemoteMultiaddr())
	if err != nil {
		return &net.TCPAddr{IP: net.ParseIP("127.1.0.1"), Port: 0}
	}
	return addr
}

var _ net.Conn = &p2pConn{}
