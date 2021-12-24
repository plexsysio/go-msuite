package grpcsvc

import (
	"context"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	gtrace "github.com/moxiaomomo/grpc-jaeger"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/plexsysio/go-msuite/modules/auth"
	"github.com/plexsysio/go-msuite/modules/config"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var JwtAuth = fx.Options(
	fx.Provide(JwtAuthOptions),
)

type JwtAuthOpts struct {
	fx.Out

	UOut grpc.UnaryServerInterceptor  `group:"unary_opts"`
	SOut grpc.StreamServerInterceptor `group:"stream_opts"`
}

func JwtAuthOptions(
	jm auth.JWTManager,
	am auth.ACL,
) (params JwtAuthOpts, err error) {
	incp := NewAuthInterceptor(jm, am)
	params.UOut = incp.Unary()
	params.SOut = incp.Stream()
	return
}

// AuthInterceptor is a server interceptor for authentication and authorization
type AuthInterceptor struct {
	jm auth.JWTManager
	am auth.ACL
}

// NewAuthInterceptor returns a new auth interceptor
func NewAuthInterceptor(jm auth.JWTManager, am auth.ACL) *AuthInterceptor {
	return &AuthInterceptor{jm, am}
}

// Unary returns a server interceptor function to authenticate and authorize unary RPC
func (interceptor *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		err := interceptor.authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// Stream returns a server interceptor function to authenticate and authorize stream RPC
func (interceptor *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		err := interceptor.authorize(stream.Context(), info.FullMethod)
		if err != nil {
			return err
		}
		return handler(srv, stream)
	}
}

func (interceptor *AuthInterceptor) authorize(ctx context.Context, method string) error {
	roles := interceptor.am.Allowed(ctx, method)
	for _, rl := range roles {
		if rl == auth.None {
			// everyone can access
			return nil
		}
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}
	values := md["authorization"]
	if len(values) == 0 {
		return status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}
	accessToken := values[0]
	claims, err := interceptor.jm.Verify(accessToken)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}
	for _, role := range roles {
		if string(role) == claims.Role {
			return nil
		}
	}
	return status.Error(codes.PermissionDenied, "no permission to access this RPC")
}

var Prometheus = fx.Options(
	fx.Provide(Metrics),
	fx.Provide(MetricsOpts),
	fx.Invoke(MetricsRegister),
)

type PrometheusOpts struct {
	fx.Out

	UOut grpc.UnaryServerInterceptor  `group:"unary_opts"`
	SOut grpc.StreamServerInterceptor `group:"stream_opts"`
}

func Metrics() *grpc_prometheus.ServerMetrics {
	return grpc_prometheus.NewServerMetrics()
}

func MetricsOpts(c config.Config, grpcMetrics *grpc_prometheus.ServerMetrics) (params PrometheusOpts, err error) {
	if c.IsSet("UsePrometheusLatency") {
		grpcMetrics.EnableHandlingTimeHistogram()
	}
	params.SOut = grpcMetrics.StreamServerInterceptor()
	params.UOut = grpcMetrics.UnaryServerInterceptor()
	return
}

func MetricsRegister(reg *prometheus.Registry, s *grpc.Server, metrics *grpc_prometheus.ServerMetrics) {
	reg.MustRegister(metrics)
	metrics.InitializeMetrics(s)
}

var TracerModule = fx.Options(
	fx.Provide(JaegerTracerOptions),
)

type TracerOpts struct {
	fx.Out

	UOut grpc.UnaryServerInterceptor `group:"unary_opts"`
}

func JaegerTracerOptions(tracer opentracing.Tracer) (params TracerOpts) {
	params.UOut = gtrace.ServerInterceptor(tracer)
	return
}
