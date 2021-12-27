package events_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	logger "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/plexsysio/go-msuite/modules/events"
	"github.com/plexsysio/taskmanager"
)

type testEvent struct {
	Msg string
}

func (testEvent) Topic() string {
	return "testEvent"
}

func (t *testEvent) Marshal() ([]byte, error) {
	return json.Marshal(t)
}

func (t *testEvent) Unmarshal(buf []byte) error {
	return json.Unmarshal(buf, t)
}

func TestEvents(t *testing.T) {
	_ = logger.SetLogLevel("pubsub", "Debug")

	tm1 := taskmanager.New(0, 2, time.Second)
	tm2 := taskmanager.New(0, 2, time.Second)

	h1, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"))
	if err != nil {
		t.Fatal(err)
	}

	h2, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"))
	if err != nil {
		t.Fatal(err)
	}

	psub1, err := pubsub.NewGossipSub(context.TODO(), h1, pubsub.WithFloodPublish(true))
	if err != nil {
		t.Fatal(err)
	}

	psub2, err := pubsub.NewGossipSub(context.TODO(), h2, pubsub.WithFloodPublish(true))
	if err != nil {
		t.Fatal(err)
	}

	ev1, err := events.NewEventsSvc(psub1, tm1)
	if err != nil {
		t.Fatal(err)
	}

	ev2, err := events.NewEventsSvc(psub2, tm2)
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

	mtx := sync.Mutex{}
	count1, count2 := 0, 0

	ev1.RegisterHandler(func() events.Event { return new(testEvent) }, func(ev events.Event) {
		testEv, ok := ev.(*testEvent)
		if !ok {
			t.Fatal("invalid event in handler")
		}
		if testEv.Msg != "hello" && testEv.Msg != "world" {
			t.Fatal("incorrect msg in event")
		}
		mtx.Lock()
		count1++
		mtx.Unlock()
	})

	ev2.RegisterHandler(func() events.Event { return new(testEvent) }, func(ev events.Event) {
		testEv, ok := ev.(*testEvent)
		if !ok {
			t.Fatal("invalid event in handler")
		}
		if testEv.Msg != "hello" && testEv.Msg != "world" {
			t.Fatal("incorrect msg in event")
		}
		mtx.Lock()
		count2++
		mtx.Unlock()
	})

	err = ev1.Broadcast(context.TODO(), &testEvent{Msg: "hello"})
	if err != nil {
		t.Fatal(err)
	}

	err = ev2.Broadcast(context.TODO(), &testEvent{Msg: "world"})
	if err != nil {
		t.Fatal(err)
	}

	started := time.Now()
	for {
		time.Sleep(time.Second)

		mtx.Lock()
		if count1 == 2 && count2 == 2 {
			mtx.Unlock()
			break
		}
		mtx.Unlock()

		if time.Since(started) > 3*time.Second {
			t.Fatal("waited 3 secs for events to trigger", count1, count2)
		}
	}

}
