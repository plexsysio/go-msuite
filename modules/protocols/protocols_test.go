package protocols_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/plexsysio/go-msuite/modules/protocols"
)

type testProtocol struct {
	sendMsg  protocols.Sender
	rcvdMsgs int
	sentMsgs int
}

type testMsg struct {
	Msg string
}

func (t *testMsg) Marshal() ([]byte, error) {
	return json.Marshal(t)
}

func (t *testMsg) Unmarshal(buf []byte) error {
	return json.Unmarshal(buf, t)
}

func (testProtocol) ID() protocol.ID {
	return protocol.ID("/testproto/1.0.0")
}

func (t *testProtocol) HandleMsg(msg protocols.Request, _ peer.ID) (protocols.Response, error) {
	t.rcvdMsgs++
	return msg, nil
}

func (t *testProtocol) SetSender(s protocols.Sender) {
	t.sendMsg = s
}

func (testProtocol) ReqFactory() protocols.Request {
	return new(testMsg)
}

func (testProtocol) RespFactory() protocols.Response {
	return new(testMsg)
}

func (t *testProtocol) Send(ctx context.Context, p peer.ID, msg protocols.Message) (protocols.Message, error) {
	t.sentMsgs++
	return t.sendMsg(ctx, p, msg)
}

func TestProtocol(t *testing.T) {

	h1, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"))
	if err != nil {
		t.Fatal(err)
	}

	h2, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"))
	if err != nil {
		t.Fatal(err)
	}

	err = h1.Connect(context.TODO(), peer.AddrInfo{
		ID:    h2.ID(),
		Addrs: h2.Addrs(),
	})
	if err != nil {
		t.Fatal(err)
	}

	svc1 := protocols.New(h1)
	svc2 := protocols.New(h2)

	p1 := &testProtocol{}
	p2 := &testProtocol{}

	svc1.Register(p1)
	svc2.Register(p2)

	resp, err := p1.Send(context.TODO(), h2.ID(), &testMsg{Msg: "Hello!"})
	if err != nil {
		t.Fatal(err)
	}

	if respMsg, ok := resp.(*testMsg); !ok {
		t.Fatal("invalid resp msg")
	} else if respMsg.Msg != "Hello!" {
		t.Fatalf("invalid resp msg recvd expected %q got %q", "Hello!", respMsg.Msg)
	}

	resp2, err := p2.Send(context.TODO(), h1.ID(), &testMsg{Msg: "World!"})
	if err != nil {
		t.Fatal(err)
	}

	if respMsg2, ok := resp2.(*testMsg); !ok {
		t.Fatal("invalid resp msg")
	} else if respMsg2.Msg != "World!" {
		t.Fatalf("invalid resp msg recvd expected %q got %q", "World!", respMsg2.Msg)
	}

	if p1.rcvdMsgs != 1 || p1.sentMsgs != 1 || p2.rcvdMsgs != 1 || p2.sentMsgs != 1 {
		t.Fatal("incorrect count of msgs")
	}
}
