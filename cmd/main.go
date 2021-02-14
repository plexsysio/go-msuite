package main

import (
	"context"
	"fmt"
	"github.com/aloknerurkar/go-msuite/modules/config"
	"github.com/aloknerurkar/go-msuite/modules/config/json"
	"github.com/aloknerurkar/go-msuite/modules/grpc"
	"github.com/aloknerurkar/go-msuite/modules/ipfs"
	"github.com/aloknerurkar/go-msuite/modules/locker"
	"github.com/aloknerurkar/go-msuite/modules/repo/fsrepo"
	ds "github.com/ipfs/go-datastore"
	logger "github.com/ipfs/go-log/v2"
	"github.com/mitchellh/go-homedir"
	"go.uber.org/fx"
	"path/filepath"
)

func main() {
	logger.SetLogLevel("*", "Debug")

	ctx, cancel := context.WithCancel(context.Background())

	hd, err := homedir.Dir()
	if err != nil {
		fmt.Println("Failed to identify home directory")
		return
	}
	rootPath := filepath.Join(hd, ".msuite")
	if !fsrepo.IsInitialized(rootPath) {
		fmt.Println("Initializing new repository at", rootPath)
		err := fsrepo.Init(rootPath, jsonConf.DefaultConfig())
		if err != nil {
			fmt.Println("Failed to initialize new repository Err:", err.Error())
			return
		}
	}
	r, err := fsrepo.Open(rootPath)
	if err != nil {
		fmt.Println("Failed to open repository Err:", err.Error())
		return
	}

	app := fx.New(
		fx.Provide(func() context.Context {
			return ctx
		}),
		fx.Provide(func() (config.Config, ds.Batching) {
			return r.Config(), r.Datastore()
		}),
		ipfs.Module,
		locker.Module,
		grpcServer.Module(r.Config()),
	)

	fmt.Println("Starting")
	err = app.Start(ctx)
	if err != nil {
		fmt.Println("Failed starting app")
		cancel()
		return
	}

	<-app.Done()
	cancel()
	_ = app.Stop(ctx)
}
