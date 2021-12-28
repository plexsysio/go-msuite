package protocols

import (
	"bufio"
	"context"
	"encoding/binary"
	"io"
	"time"

	logger "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
)

const defaultTimeout = 15 * time.Second

var log = logger.Logger("protocols")

// ProtocolsSvc provides an easier interface to write P2P protocols in terms of
// request-response schemes
type ProtocolsSvc interface {
	Register(Protocol)
}

// Message type is a generic type which is serializable
type Message interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

// Request can be any type that satisfies the Message interface
type Request Message

// Response can be any type that satisfies the Message interface
type Response Message

// Sender is an interface used to provide message sending functionality to the protocol
// Protocols which register are provided the Sender functionalityto use
type Sender func(context.Context, peer.ID, Request) (Response, error)

// Protocol interface defines contract to implement to have a new P2P protocol registered
// on the host
type Protocol interface {
	// ID returns the protocol Identifier as specified in the libp2p-core pkg
	ID() protocol.ID
	// HandleMsg is used to handle new incoming request on the host. The response is sent
	// back on the same stream
	HandleMsg(Request, peer.ID) (Response, error)
	// SetSender will be used to provide protocols the ability to send messages. It will
	// provide an interface which can be used to send the messages on the wire. This sender
	// can be used within the protocol to initiate message sending
	SetSender(Sender)
	// ReqFactory is used to instantiate a new object for handling the request
	ReqFactory() Request
	// RespFactory is used to instantiate a new object for handling the response
	RespFactory() Response
}

type service struct {
	h host.Host
}

func New(h host.Host) ProtocolsSvc {
	return &service{h}
}

func (s *service) Register(p Protocol) {
	s.h.SetStreamHandler(p.ID(), func(stream network.Stream) {

		_ = stream.SetDeadline(time.Now().Add(defaultTimeout))

		rdr, wrtr := newReader(stream), newWriter(stream)

		req := p.ReqFactory()

		err := rdr.ReadMsg(req)
		if err != nil {
			_ = stream.Reset()
			log.Error("failed reading stream", err)
			return
		}

		resp, err := p.HandleMsg(req, stream.Conn().RemotePeer())
		if err != nil {
			_ = stream.Reset()
			log.Error("HandleMsg returned error", err)
			return
		}

		err = wrtr.WriteMsg(resp)
		if err != nil {
			_ = stream.Reset()
			log.Error("failed to write msg on wire", err)
			return
		}
	})
	p.SetSender(func(ctx context.Context, id peer.ID, req Request) (Response, error) {
		return s.SendMsg(ctx, id, p, req)
	})
}

func (s *service) SendMsg(ctx context.Context, id peer.ID, p Protocol, msg Message) (Message, error) {
	stream, err := s.h.NewStream(ctx, id, p.ID())
	if err != nil {
		log.Error("failed opening new stream to peer", id, p)
		return nil, err
	}

	defer stream.Close()

	_ = stream.SetDeadline(time.Now().Add(defaultTimeout))

	rdr, wrtr := newReader(stream), newWriter(stream)

	err = wrtr.WriteMsg(msg)
	if err != nil {
		log.Error("failed writing message", err)
		return nil, err
	}

	resp := p.RespFactory()

	err = rdr.ReadMsg(resp)
	if err != nil {
		log.Error("failed reading from stream", err)
		return nil, err
	}

	return resp, nil
}

type msgReader struct {
	r io.Reader
}

func newReader(stream network.Stream) *msgReader {
	return &msgReader{r: bufio.NewReader(stream)}
}

func (m *msgReader) ReadMsg(msg Message) error {
	lengthBuf := make([]byte, 2)
	_, err := m.r.Read(lengthBuf)
	if err != nil {
		return err
	}

	length := binary.LittleEndian.Uint16(lengthBuf)

	buf := make([]byte, length)
	_, err = io.ReadFull(m.r, buf)
	if err != nil {
		return err
	}

	return msg.Unmarshal(buf)
}

type msgWriter struct {
	w io.Writer
}

func newWriter(stream network.Stream) *msgWriter {
	return &msgWriter{w: stream}
}

func (m *msgWriter) WriteMsg(msg Message) error {
	msgBuf, err := msg.Marshal()
	if err != nil {
		return err
	}

	lenBuf := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenBuf, uint16(len(msgBuf)))

	_, err = m.w.Write(lenBuf)
	if err != nil {
		return err
	}

	_, err = m.w.Write(msgBuf)
	return err
}
