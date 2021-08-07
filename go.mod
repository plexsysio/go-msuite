module github.com/plexsysio/go-msuite

go 1.13

require (
	github.com/SWRMLabs/ants-db v0.0.3
	github.com/SWRMLabs/ss-ds-store v0.0.7
	github.com/SWRMLabs/ss-store v0.0.4
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/golang/snappy v0.0.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.3.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hsanjuan/ipfs-lite v1.1.18
	github.com/ipfs/go-blockservice v0.1.4
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-datastore v0.4.5
	github.com/ipfs/go-ds-flatfs v0.4.5
	github.com/ipfs/go-ds-leveldb v0.4.2
	github.com/ipfs/go-ipfs-blockstore v1.0.3
	github.com/ipfs/go-ipld-format v0.2.0
	github.com/ipfs/go-log/v2 v2.3.0
	github.com/ipfs/go-merkledag v0.3.2
	github.com/libp2p/go-libp2p v0.13.0
	github.com/libp2p/go-libp2p-connmgr v0.2.4
	github.com/libp2p/go-libp2p-core v0.8.5
	github.com/libp2p/go-libp2p-discovery v0.5.0
	github.com/libp2p/go-libp2p-gostream v0.3.1
	github.com/libp2p/go-libp2p-pubsub v0.4.1
	github.com/libp2p/go-libp2p-quic-transport v0.10.0
	github.com/libp2p/go-libp2p-tls v0.1.3
	github.com/mitchellh/go-homedir v1.1.0
	github.com/moxiaomomo/grpc-jaeger v0.0.0-20180617090213-05b879580c4a
	github.com/multiformats/go-multiaddr v0.3.1
	github.com/opentracing/opentracing-go v1.2.0
	github.com/plexsysio/dLocker v0.0.2
	github.com/plexsysio/taskmanager v0.0.0-20210719193446-5b3bff8bc055
	github.com/prometheus/client_golang v1.11.0
	github.com/rs/cors v1.7.0
	github.com/slok/go-http-metrics v0.9.0
	github.com/uber/jaeger-client-go v2.22.1+incompatible // indirect
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	go.uber.org/fx v1.10.0
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.18.1 // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	google.golang.org/grpc v1.36.0
)

replace github.com/hsanjuan/ipfs-lite => github.com/plexsysio/ipfs-lite v1.1.21

replace github.com/SWRMLabs/ants-db => github.com/plexsysio/ants-db v0.0.4
