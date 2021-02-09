package main

import (
	"context"
	"fmt"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/aloknerurkar/go-msuite/modules/grpc"
	"github.com/aloknerurkar/go-msuite/modules/ipfs"
	logger "github.com/ipfs/go-log"
	"go.uber.org/fx"
)

func main() {
	logger.SetLogLevel("*", "Debug")
	app := fx.New(
		config.Module,
		ipfs.P2P,
		grpcServer.JwtAuth,
		grpcServer.TracerModule,
		grpcServer.Module,
	)

	ctx := context.Background()

	fmt.Println("Starting")
	err := app.Start(ctx)
	if err != nil {
		fmt.Println("Failed starting app")
		return
	}

	<-app.Done()
	_ = app.Stop(ctx)
}
