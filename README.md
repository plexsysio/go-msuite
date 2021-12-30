# go-msuite [![Go](https://github.com/plexsysio/go-msuite/workflows/Go/badge.svg)](https://github.com/plexsysio/go-msuite/actions) [![Go Reference](https://pkg.go.dev/badge/github.com/plexsysio/go-msuite.svg)](https://pkg.go.dev/github.com/plexsysio/go-msuite) [![Coverage Status](https://coveralls.io/repos/github/plexsysio/go-msuite/badge.svg?branch=master)](https://coveralls.io/github/plexsysio/go-msuite?branch=master) [![DeepSource](https://deepsource.io/gh/plexsysio/go-msuite.svg/?label=active+issues&show_trend=true&token=ObURascsEDWqoJfshvTBOc-w)](https://deepsource.io/gh/plexsysio/go-msuite/?ref=repository-badge)
Modular microservices framework in golang

## Introduction

`go-msuite` aims to help bootstrap development of distributed applications in go. It does not add too many new components, but it a collection of a bunch of useful software already built and battle-tested. There are a bunch of defaults already present, however the library is built in such a way that we can plug-n-play different subsystems.

Features are mostly similar to other popular frameworks. The only difference is `go-msuite` uses [libp2p](https://github.com/libp2p) underneath. Libp2p is a modular networking stack built for p2p applications. With Libp2p, we have the ability to add more transports underneath like QUIC, websockets etc. Apart from this, we get access to the DHT, which provides routing, discovery and even pubsub functionality.

For the public API please go through the `core` package

## Features

- Service lifecycle
	- `go-msuite` uses [uber/fx](go.uber.org/fx) for dependency injection and lifecycle management. This provides a simple Start/Stop type interface to developers to manage their apps

- HTTP and gRPC endpoint
   - Most of the applications today use HTTP or RPC interface. gRPC being very popular and having a very broad ecosystem. `go-msuite` takes care of the lifecycle of your HTTP and gRPC servers, which can be used to register services/endpoints.
   - Naturally, a bunch of middlewares are implemented to take care of auth, tracing, metrics etc. This is again common stuff which needs to be re-implemented each time an application is built.

- Libp2p and IPFS
   - A libp2p host is instantiated by `go-msuite`. It is possible to use existing keys or create new ones. Each application has access to [libp2p-host](https://github.com/libp2p/go-libp2p-core/tree/master/host) and hence all the functionality that goes with it.
   - [ipfs-lite](https://github.com/hsanjuan/ipfs-lite) is instantiated using the above libp2p host and the [repository storage](https://github.com/plexsysio/go-msuite/tree/master/modules/repo). This can be used to share data between different services in the form of files.
   - [Pubsub](https://github.com/libp2p/go-libp2p-pubsub) and [Discovery](https://github.com/libp2p/go-libp2p-discovery) are also supported using libp2p.

- RPC Transport
   - There are multiple transports available. Users can start `go-msuite` using just TCP/UDS transport as well, libp2p is completely optional. That said, gRPC services registered on `go-msuite` are available on all the transports that are configured.
   - Users can configure ports for different transports

- Authentication
   - Authentication is added as first-class citizen. Currently a JWT-based implementation exists. User can enable it by providing a secret phase.
   - Access control is also present. Users can configure their APIs/Services to start using ACLs. These can be updated/removed etc.

- Storage and SharedStorage
   - Currently a simple key-value store is available to all the services. This store uses a very generic K-V store interface [gkvstore](https://github.com/plexsysio/gkvstore) which allows users to define how they want to store the objects into the store. Different implementations can be added here in future.
   - Using IPFS and the above storage, we can use [ants-db](https://github.com/plexsysio/ants-db) to have a distributed CRDT store across all the go-msuite instances. This is particularly useful to share state across all the instances running. Also it piggy-backs on the existing libp2p and IPFS instances that we already configure.

- Taskmanager
	- This is a simple worker pool which can be used to run tasks asynchronously. [taskmanager](http://github.com/plexsysio/taskmanager) can be configured to have dedicated go-routines which handle tasks created by users. Tasks can be short/long. All the tasks are stopped on app close. Also there is way to show task progress on the diagnostic endpoint

- Distributed locking
   - Distributed locking is useful when you have multiple instances of your services running. This way we can synchronize services across different machines. This component currently uses zookeeper/redis implementations which need to be managed separately.

- Events
   - Simple event framework over libp2p Pubsub. This can be used inside apps to create and react to events and is configurable by users. It provides an easier message-based interface to send/react to things. The events should be idempotent as they can be fired on multiple receivers.

- Protocols
   - Protocols service can be used to write request-response schemes over libp2p. This allows users to write libp2p protocols with a simple message-passing model. There is a protocol internally implemented to provide a naive service mesh functionality.

- Diagnostics
   - HTTP endpoint for showing diagnostic information of `go-msuite`. This shows status of different servers and routines started by the user. Users can add information to this and observe it from the HTTP interface. If the routines are created using `taskmanager`, the `Status` interface of `taskmanager` can be used to print status of the workers on the HTTP endpoints.
   - `pprof` HTTP handlers can be enabled for debugging
   - `prometheus` HTTP handler can also be enabled if metrics is enabled. It should be possible to use the same registry to add metrics in user apps.
   - `opentracing-tracer` can be configured. Both the gRPC services and HTTP services will be able to use this. Additionally user can access the tracer to add more custom traces.

- Service discovery
   - Each `go-msuite` instance or individual service can be started with a particular name. This name can be then used to connect to it from other `go-msuite` nodes. Currently, it uses libp2p discovery underneath as mentioned above.
   - A static configuration is also possible of the nodes and IP addresses are known in advance and libp2p is not configured.

## Quickstart


```
	import github.com/plexsysio/go-msuite

	svc, err := msuite.New(
		msuite.WithServices("Hello", "World"),
		msuite.WithHTTP(8080),
		msuite.WithP2P(10000),
		msuite.WithGRPC("tcp", 10001),
		msuite.WithGRPC("p2p", nil),
	)

	// write your app which uses the different subsystems from svc
	err = yourApp.NewService(svc)
	if err != nil {
		fmt.Println("failed creating new service", err.Error())
		return
	}
	err = svc.Start(context.Background())
	if err != nil {
		fmt.Println("failed starting service", err.Error())
		return
	}
	<-svc.Done()
	svc.Stop(context.Background())
```

## Examples
There is a separate [repository](https://github.com/plexsysio/msuite-services) which contains different services built using `go-msuite`.

## License
MIT licensed

