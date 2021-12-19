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
	"github.com/plexsysio/go-msuite/modules/repo"
	"github.com/plexsysio/go-msuite/modules/sharedStorage"
	"github.com/plexsysio/taskmanager"
	"google.golang.org/grpc"
)

type Service interface {
	Start(context.Context) error
	Stop(context.Context) error
	Done() <-chan os.Signal

	Repo() repo.Repo
	Auth() Auth
	TM() (*taskmanager.TaskManager, error)
	Node() (Node, error)
	GRPC() (GRPC, error)
	HTTP() (HTTP, error)
	Locker() (dLocker.DLocker, error)
	Events() (events.Events, error)
	SharedStorage(string, sharedStorage.Callback) (store.Store, error)
	Files() (*ipfslite.Peer, error)
}

type Node interface {
	P2P() P2P
	Pubsub() *pubsub.PubSub
}

type P2P interface {
	Host() host.Host
	Routing() routing.Routing
	Discovery() discovery.Discovery
}

type Auth interface {
	JWT() (auth.JWTManager, error)
	ACL() (auth.ACL, error)
}

type GRPC interface {
	Server() *grpc.Server
	Client(context.Context, string, ...grpc.DialOption) (*grpc.ClientConn, error)
}

type HTTP interface {
	Mux() *http.ServeMux
	Gateway() *runtime.ServeMux
}
