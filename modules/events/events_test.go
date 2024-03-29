package events_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	logger "github.com/ipfs/go-log/v2"
	bhost "github.com/libp2p/go-libp2p-blankhost"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	swarmt "github.com/libp2p/go-libp2p-swarm/testing"
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

	h1 := bhost.NewBlankHost(swarmt.GenSwarm(t, swarmt.OptDisableQUIC))
	h2 := bhost.NewBlankHost(swarmt.GenSwarm(t, swarmt.OptDisableQUIC))

	t.Cleanup(func() {
		tm1.Stop()
		tm2.Stop()
		h1.Close()
		h2.Close()
	})

	psub1, err := pubsub.NewFloodSub(context.TODO(), h1)
	if err != nil {
		t.Fatal(err)
	}

	psub2, err := pubsub.NewFloodSub(context.TODO(), h2)
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
		// Event can be fired on both nodes or on only 1
		if count1 > 0 && count2 > 0 {
			mtx.Unlock()
			break
		}
		mtx.Unlock()

		if time.Since(started) > 3*time.Second {
			t.Fatal("waited 3 secs for events to trigger", count1, count2)
		}
	}

}
