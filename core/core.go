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
	"github.com/opentracing/opentracing-go"
	"github.com/plexsysio/dLocker"
	store "github.com/plexsysio/gkvstore"
	"github.com/plexsysio/go-msuite/modules/auth"
	"github.com/plexsysio/go-msuite/modules/events"
	"github.com/plexsysio/go-msuite/modules/protocols"
	"github.com/plexsysio/go-msuite/modules/repo"
	"github.com/plexsysio/go-msuite/modules/sharedStorage"
	"github.com/plexsysio/taskmanager"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

// Service is the collection of all the interfaces provided by the msuite instance
type Service interface {
	Start(context.Context) error
	Stop(context.Context) error
	Done() <-chan os.Signal

	// Following packages are must
	// Repo manages the on-disk/inmem state of the application
	Repo() repo.Repo
	// TM uses taskmanager for async task scheduling within
	TM() *taskmanager.TaskManager

	// Auth is used to provide authorized access to resources
	Auth() (Auth, error)
	// P2P encapsulates the libp2p related functionality which can be optionally
	// configured and used
	P2P() (P2P, error)
	// GRPC encapsulates the gRPC client-server functionalities
	GRPC() (GRPC, error)
	// HTTP provides a multiplexer with middlewares configured. Also it provides
	// the gRPC Gateway mux registered on the main mux
	HTTP() (HTTP, error)
	// Locker provides access to the distributed locker configured if any
	Locker() (dLocker.DLocker, error)
	// Events service can be used to broadcast/handle events in the form of messages
	// using underlying PubSub
	Events() (events.Events, error)
	// Protocols service provides a simple request-response protocol interface to use
	Protocols() (protocols.ProtocolsSvc, error)
	// SharedStorage provides access to a distributed CRDT K-V store. Callbacks can
	// be registered to get updates about certain keys
	SharedStorage(string, sharedStorage.Callback) (store.Store, error)
	// Files gives access to the ipfslite.Peer object. This can be used to share
	// files across different nodes
	Files() (*ipfslite.Peer, error)
	// Tracing provides access to the configured tracer
	Tracing() (opentracing.Tracer, error)
	// Metrics provides access to the prometheus registry. This registry already
	// has a bunch of default metrics registered.
	Metrics() (*prometheus.Registry, error)
}

// P2P encapsulates the libp2p functionality. These can be used to write more
// advanced protocols/features if required
type P2P interface {
	Host() host.Host
	Routing() routing.Routing
	Discovery() discovery.Discovery
	Pubsub() *pubsub.PubSub
}

// Auth provides authorized access to resources using ACLs and JWT tokens
type Auth interface {
	JWT() auth.JWTManager
	ACL() auth.ACL
}

// GRPC provides the gRPC client-server implementations. Can be used to register services
// or call other services already registered
type GRPC interface {
	Server() *grpc.Server
	Client(context.Context, string, ...grpc.DialOption) (*grpc.ClientConn, error)
}

// HTTP provides the standard HTTP multiplexer already configured with middlewares. This
// can be used to register handlers etc. Gateway provides GRPC gateway multiplexer which
// can be used to register gRPC-gateways
type HTTP interface {
	Mux() *http.ServeMux
	Gateway() *runtime.ServeMux
}
