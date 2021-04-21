package msuite

import (
	"context"
	"github.com/SWRMLabs/ss-store"
	"github.com/SWRMLabs/ss-taskmanager"
	"github.com/plexsysio/dLocker"
	"github.com/plexsysio/go-msuite/modules/auth"
	"github.com/plexsysio/go-msuite/modules/events"
	"github.com/plexsysio/go-msuite/modules/repo"
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

	Repo() repo.Repo
	Auth() Auth
	TM() (*taskmanager.TaskManager, error)
	Node() (Node, error)
	GRPC() (GRPC, error)
	HTTP() (HTTP, error)
	Locker() (dLocker.DLocker, error)
	Events() (events.Events, error)
}

type Node interface {
	Storage() store.Store
	P2P() P2P
	Pubsub() *pubsub.PubSub
	IPFS() *ipfslite.Peer
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
	Client(context.Context, string) (*grpc.ClientConn, error)
}

type HTTP interface {
	Mux() *http.ServeMux
}
