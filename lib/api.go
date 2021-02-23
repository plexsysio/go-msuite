package msuite

import (
	"context"
	"github.com/StreamSpace/ss-store"
	"github.com/aloknerurkar/dLocker"
	"github.com/aloknerurkar/go-msuite/modules/auth"
	"github.com/aloknerurkar/go-msuite/modules/events"
	"github.com/aloknerurkar/go-msuite/modules/grpc/client"
	"github.com/aloknerurkar/go-msuite/modules/repo"
	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/routing"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"google.golang.org/grpc"
	"net/http"
	"os"
)

type Service interface {
	Start(context.Context) error
	Stop(context.Context) error
	Done() <-chan os.Signal

	Node() Node
	Auth() Auth
	GRPC() GRPC
	HTTP() HTTP
	Locker() dLocker.DLocker
	Events() events.Events
}

type Node interface {
	Repo() repo.Repo
	Storage() Storage
	P2P() P2P
	Pubsub() *pubsub.PubSub
	IPFS() *ipfslite.Peer
}

type P2P interface {
	Host() host.Host
	Routing() routing.Routing
	Discovery() discovery.Discovery
}

type Storage interface {
	Local() store.Store
	Shared() store.Store
}

type Auth interface {
	JWT() auth.JWTManager
	ACL() auth.ACL
}

type GRPC interface {
	Server() *grpc.Server
	Client(context.Context, string) (grpcclient.Client, error)
	// Gateway() *runtime.ServeMux
}

type HTTP interface {
	Mux() *http.ServeMux
}
