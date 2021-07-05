package grpcmux_test

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/SWRMLabs/ss-taskmanager"
	"github.com/plexsysio/go-msuite/modules/grpc/transport/mux"
)

func TestMultipleListeners(t *testing.T) {
	tcpListener1, _ := net.Listen("tcp", ":8080")
	tcpListener2, _ := net.Listen("tcp", ":8081")
	tcpListener3, _ := net.Listen("tcp", ":8082")
	tm := taskmanager.NewTaskManager(context.Background(), 4)

	listeners := grpcmux.MuxIn{
		Listeners: []grpcmux.MuxListener{
			{
				Listener: tcpListener1,
				Tag:      "1",
			},
			{
				Listener: tcpListener2,
				Tag:      "2",
			},
			{
				Listener: tcpListener3,
				Tag:      "3",
			},
		},
	}

	m := grpcmux.New(
		context.Background(),
		listeners,
		tm,
	)

	m.Start(nil)

	connChan := make(chan net.Conn)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(connChan)
		for {
			conn, err := m.Accept()
			if err != nil {
				return
			}
			connChan <- conn
		}
	}()

	count := 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _ = range connChan {
			count++
			if count == 3 {
				m.Close()
			}
		}
	}()

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	_, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Fatal(err)
	}
	_, err = net.Dial("tcp", ":8081")
	if err != nil {
		t.Fatal(err)
	}
	_, err = net.Dial("tcp", ":8082")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-done:
	case <-time.After(time.Second * 3):
		t.Fatal("waited 3 secs for done")
	}
}
