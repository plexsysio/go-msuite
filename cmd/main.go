package main

import (
	"context"
	"fmt"
	"github.com/aloknerurkar/go-msuite/modules/config/json"
	"github.com/aloknerurkar/go-msuite/modules/grpc"
	"github.com/aloknerurkar/go-msuite/modules/ipfs"
	"github.com/aloknerurkar/go-msuite/modules/locker"
	"github.com/aloknerurkar/go-msuite/modules/repo/fsrepo"
	logger "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"
)

func main() {
	logger.SetLogLevel("*", "Debug")
	app := fx.New(
		jsonConf.Default,
		fsrepo.Module,
		ipfs.Module,
		locker.Module,
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
