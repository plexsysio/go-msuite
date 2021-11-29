package grpcmux_test

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	logger "github.com/ipfs/go-log/v2"
	grpcmux "github.com/plexsysio/go-msuite/modules/grpc/mux"
	"github.com/plexsysio/taskmanager"
)

func TestMultipleListeners(t *testing.T) {
	_ = logger.SetLogLevel("grpc/lmux", "*")

	tcpListener1, err := net.Listen("tcp", ":10081")
	if err != nil {
		t.Fatal(err)
	}
	tcpListener2, err := net.Listen("tcp", ":10082")
	if err != nil {
		t.Fatal(err)
	}
	tcpListener3, err := net.Listen("tcp", ":10083")
	if err != nil {
		t.Fatal(err)
	}

	tm := taskmanager.New(4, 10, time.Second*10)

	listeners := []grpcmux.MuxListener{
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
	}

	m := grpcmux.New(
		context.Background(),
		listeners,
		tm,
	)

	m.Start(context.TODO(), nil)

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
		for range connChan {
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

	_, err = net.Dial("tcp", ":10081")
	if err != nil {
		t.Fatal(err)
	}
	_, err = net.Dial("tcp", ":10082")
	if err != nil {
		t.Fatal(err)
	}
	_, err = net.Dial("tcp", ":10083")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-done:
	case <-time.After(time.Second * 3):
		t.Fatal("waited 3 secs for done")
	}
}
