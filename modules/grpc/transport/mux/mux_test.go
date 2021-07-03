package grpcmux_test

import (
	"context"
	"net"
	"testing"

	"github.com/SWRMLabs/ss-taskmanager"
	"github.com/plexsysio/go-msuite/modules/grpc/transport/mux"
)

func TestMultipleListeners(t *testing.T) {
	tcpListener1, _ := net.Listen("tcp", ":8080")
	tcpListener2, _ := net.Listen("tcp", ":8081")
	tcpListener3, _ := net.Listen("tcp", ":8082")
	tm := taskmanager.NewTaskManager(context.Background(), 4)

	listeners := grpcmux.MuxIn{
		Listeners: {
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
}
