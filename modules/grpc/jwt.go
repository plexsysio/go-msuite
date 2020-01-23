package grpcServer

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/dgrijalva/jwt-go"
	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var JwtAuth = fx.Options(
	fx.Provide(JwtAuthOptions),
)

type JwtAuthOpts struct {
	fx.Out

	UOut grpc.UnaryServerInterceptor  `group:"unary_opts"`
	SOut grpc.StreamServerInterceptor `group:"stream_opts"`
}

func JwtAuthOptions(conf config.Config) (params JwtAuthOpts, err error) {

	useJwt, ok := conf.Get("use_jwt").(bool)
	if !ok {
		useJwt = false
	}
	if useJwt {
		pubKey, ok := conf.Get("pub_key").(*rsa.PublicKey)
		if !ok {
			err = errors.New("Public key not specified")
			return
		}
		log.Infof("Registering JWT Auth options")
		params.UOut = jwtUnary(pubKey)
		params.SOut = jwtStream(pubKey)
		return
	}
	return
}

func jwtUnary(publicKey *rsa.PublicKey) grpc.UnaryServerInterceptor {
	return grpc_auth.UnaryServerInterceptor(authFunc(publicKey))
}

func jwtStream(publicKey *rsa.PublicKey) grpc.StreamServerInterceptor {
	return grpc_auth.StreamServerInterceptor(authFunc(publicKey))
}

func authFunc(pubKey *rsa.PublicKey) func(context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("Metadata not present")
		}

		jwtToken, ok := md["authorization"]
		if !ok {
			return nil, errors.New("Authorization header not present")
		}

		token, err := validateToken(jwtToken[0], pubKey)
		if err != nil {
			return nil, errors.New("Invalid token")
		}

		newCtx := context.WithValue(ctx, "jwt_token", token)
		return newCtx, nil
	}
}

func validateToken(token string, publicKey *rsa.PublicKey) (*jwt.Token, error) {
	jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			log.Errorf("Unexpected signing method: %v", t.Header["alg"])
			return nil, fmt.Errorf("Invalid token %s", token)
		}
		return publicKey, nil
	})
	if err == nil && jwtToken.Valid {
		return jwtToken, nil
	}
	return nil, err
}
