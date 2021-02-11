package mware

import (
	"context"
	"errors"
	"fmt"
	"github.com/StreamSpace/ss-store"
	"github.com/aloknerurkar/go-msuite/modules/acl"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/dgrijalva/jwt-go"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"time"
)

var JwtAuth = fx.Options(
	fx.Provide(JwtAuthOptions),
)

type JwtAuthOpts struct {
	fx.Out

	JM *JWTManager

	UOut grpc.UnaryServerInterceptor  `group:"unary_opts"`
	SOut grpc.StreamServerInterceptor `group:"stream_opts"`
}

func JwtAuthOptions(
	conf config.Config,
	st store.Store,
) (params JwtAuthOpts, err error) {
	var jwtSecret string
	ok := conf.Get("JWTSecret", &jwtSecret)
	if !ok {
		err = errors.New("JWT Secret not provided")
		return
	}
	jm := NewJWTManager(jwtSecret)
	incp := NewAuthInterceptor(jm, st)

	params.JM = jm
	params.UOut = incp.Unary()
	params.SOut = incp.Stream()
	return
}

// JWTManager is a JSON web token manager
type JWTManager struct {
	secretKey string
}

// UserClaims is a custom JWT claims that contains some user's information
type UserClaims struct {
	jwt.StandardClaims
	ID   string `json:"id"`
	Role string `json:"role"`
}

type User interface {
	ID() string
	Role() string
}

// NewJWTManager returns a new JWT manager
func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{secretKey}
}

// Generate generates and signs a new token for a user
func (manager *JWTManager) Generate(user User, timeout time.Duration) (string, error) {
	claims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(timeout).Unix(),
		},
		ID:   user.ID(),
		Role: user.Role(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(manager.secretKey))

}

// Verify verifies the access token string and return a user claim if the token is valid
func (manager *JWTManager) Verify(accessToken string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}
			return []byte(manager.secretKey), nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

// AuthInterceptor is a server interceptor for authentication and authorization
type AuthInterceptor struct {
	jwtManager *JWTManager
	st         store.Store
}

// NewAuthInterceptor returns a new auth interceptor
func NewAuthInterceptor(jwtManager *JWTManager, st store.Store) *AuthInterceptor {
	return &AuthInterceptor{jwtManager, st}
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
	m := &acl.MethodRoles{
		Method: method,
	}
	err := interceptor.st.Read(m)
	if err != nil {
		// everyone can access
		return nil
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
	claims, err := interceptor.jwtManager.Verify(accessToken)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}
	for _, role := range m.Roles {
		if role == claims.Role {
			return nil
		}
	}
	return status.Error(codes.PermissionDenied, "no permission to access this RPC")
}
