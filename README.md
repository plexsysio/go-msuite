# go-msuite [![Go](https://github.com/plexsysio/go-msuite/workflows/Go/badge.svg)](https://github.com/plexsysio/go-msuite/actions) [![Go Reference](https://pkg.go.dev/badge/github.com/plexsysio/go-msuite.svg)](https://pkg.go.dev/github.com/plexsysio/go-msuite) [![Coverage Status](https://coveralls.io/repos/github/plexsysio/go-msuite/badge.svg?branch=master)](https://coveralls.io/github/plexsysio/go-msuite?branch=master) [![DeepSource](https://deepsource.io/gh/plexsysio/go-msuite.svg/?label=active+issues&show_trend=true&token=ObURascsEDWqoJfshvTBOc-w)](https://deepsource.io/gh/plexsysio/go-msuite/?ref=repository-badge)
Modular microservices framework in golang

## Introduction

`go-msuite` aims to help bootstrap development of distributed applications in go. It does not add too many new components, but it a collection of a bunch of useful software already built and battle-tested. There are a bunch of defaults already present, however the library is built in such a way that we can plug-n-play different subsystems.

Features are mostly similar to other popular frameworks. The only difference is `go-msuite` uses [libp2p](https://github.com/libp2p) underneath. Libp2p is a modular networking stack built for p2p applications. Having worked on it for more than a year now, I think it would add immense value not just to web3 but to traditional applications as well. This allows us to add more transports like websockets, quic etc.

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
   - There are multiple transports available. Users can start `go-msuite` using just TCP transport as well. libp2p is completely optional. That said, gRPC services registered on `go-msuite` are available on all the transports that are configured. (TCP and libp2p for now)
   - Users can configure ports for different transports

- Authentication
   - Authentication is added as first-class citizen. Currently a JWT-based implementation exists. User can enable it by providing a secret phase.
   - Access control is also present. Users can configure their APIs/Services to start using ACLs. These can be updated/removed etc.

- Storage and SharedStorage
   - Currently a simple key-value store is available to all the services. This store uses a very simple ORM-like interface [ss-store](https://github.com/SWRMLabs/ss-store) which allows users to define how they want to store the objects into the store. Different implementations can be added here in future.
   - Using IPFS and the above storage, we can use [ants-db](https://github.com/plexsysio/ants-db) to have a distributed CRDT store across all the go-msuite instances which are configured to use it. This is particularly useful to share state across all the instances running. Also it piggy-backs on the existing libp2p and IPFS instances that we already configure.

- Taskmanager
	- This is a simple worker pool which can be used to run tasks asynchronously. [taskmanager](http://github.com/SWRMLabs/ss-taskmanager) can be configured to have dedicated go-routines which handle tasks created by users. Tasks can be short/long. All the tasks are stopped on app close. Also there is way to show task progress on the diagnostic endpoint

- Distributed locking
   - Distributed locking is useful when you have multiple instances of your services running. This way we can synchronize services across different machines. This component currently uses zookeeper/redis implementations which need to be managed separately.

- Events
   - Simple event-based framework over libp2p Pubsub. This can be used inside apps to create and react to events configured by users. Currently, there is no synchronization, so if multiple services are present, synchronization should be done by the user. Hopefully this will change.

- Diagnostics
   - HTTP endpoint for showing diagnostic information of `go-msuite`. This shows status of different servers and routines started by the user. Users can add information to this and observe it from the HTTP interface.

- Service discovery
   - Each `go-msuite` instance can be started with a particular name. This name can be then used to connect to it from other `go-msuite` nodes. Currently, it uses libp2p discovery underneath as mentioned above.
   - A static configuration is also possible of the nodes and IP addresses are known in advance.

## Quickstart


```
	import github.com/plexsysio/go-msuite

	svc, err := msuite.New(
		msuite.WithServiceName("HelloWorld"),
		msuite.WithHTTP(8080),
		msuite.WithP2PPort(10000),
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
There is a separate [repository](https://github.com/plexsysio/msuite-services) which contains different services built using `go-msuite`. Different services use different features.

## License
MIT licensed

