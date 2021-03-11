module github.com/aloknerurkar/go-msuite

go 1.13

require (
	github.com/StreamSpace/ants-db v0.0.2
	github.com/StreamSpace/ss-ds-store v0.0.4
	github.com/StreamSpace/ss-store v0.0.2
	github.com/StreamSpace/ss-taskmanager v0.0.2
	github.com/aloknerurkar/dLocker v0.0.1
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.3.0
	github.com/hsanjuan/ipfs-lite v1.1.18
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-datastore v0.4.5
	github.com/ipfs/go-ds-badger2 v0.1.0
	github.com/ipfs/go-ds-flatfs v0.4.5
	github.com/ipfs/go-ds-leveldb v0.4.2
	github.com/ipfs/go-log/v2 v2.1.1
	github.com/libp2p/go-libp2p v0.13.0
	github.com/libp2p/go-libp2p-core v0.8.0
	github.com/libp2p/go-libp2p-discovery v0.5.0
	github.com/libp2p/go-libp2p-pubsub v0.4.1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/moxiaomomo/grpc-jaeger v0.0.0-20180617090213-05b879580c4a
	github.com/multiformats/go-multiaddr v0.3.1
	github.com/multiformats/go-multiaddr-net v0.2.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/prometheus/client_golang v1.7.1
	github.com/rs/cors v1.7.0
	github.com/slok/go-http-metrics v0.9.0
	github.com/uber/jaeger-client-go v2.22.1+incompatible // indirect
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	go.uber.org/fx v1.10.0
	google.golang.org/grpc v1.36.0
)

replace github.com/StreamSpace/ants-db => ../ants-db
