// Package core defines the core abstractions provided by msuite library
package core

import (
	"context"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/routing"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/plexsysio/dLocker"
	store "github.com/plexsysio/gkvstore"
	"github.com/plexsysio/go-msuite/modules/auth"
	"github.com/plexsysio/go-msuite/modules/events"
	"github.com/plexsysio/go-msuite/modules/protocols"
	"github.com/plexsysio/go-msuite/modules/repo"
	"github.com/plexsysio/go-msuite/modules/sharedStorage"
	"github.com/plexsysio/taskmanager"
	"google.golang.org/grpc"
)

type Service interface {
	Start(context.Context) error
	Stop(context.Context) error
	Done() <-chan os.Signal

	// Following packages are must
	// Repo manages the on-disk/inmem state of the application
	Repo() repo.Repo
	// TM uses taskmanager for async task scheduling within
	TM() *taskmanager.TaskManager

	Auth() (Auth, error)
	P2P() (P2P, error)
	GRPC() (GRPC, error)
	HTTP() (HTTP, error)
	Locker() (dLocker.DLocker, error)
	Events() (events.Events, error)
	Protocols() (protocols.ProtocolsSvc, error)
	SharedStorage(string, sharedStorage.Callback) (store.Store, error)
	Files() (*ipfslite.Peer, error)
}

type P2P interface {
	Host() host.Host
	Routing() routing.Routing
	Discovery() discovery.Discovery
	Pubsub() *pubsub.PubSub
}

type Auth interface {
	JWT() auth.JWTManager
	ACL() auth.ACL
}

type GRPC interface {
	Server() *grpc.Server
	Client(context.Context, string, ...grpc.DialOption) (*grpc.ClientConn, error)
}

type HTTP interface {
	Mux() *http.ServeMux
	Gateway() *runtime.ServeMux
}
