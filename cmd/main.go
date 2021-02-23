package main

import (
	"context"
	"fmt"
	"github.com/aloknerurkar/go-msuite/lib"
	logger "github.com/ipfs/go-log/v2"
)

func main() {
	logger.SetLogLevel("*", "Debug")

	ctx, cancel := context.WithCancel(context.Background())

	app, err := msuite.New(ctx)
	if err != nil {
		fmt.Println("Failed creating msuite service")
	}
	fmt.Println("Starting")
	err = app.Start(ctx)
	if err != nil {
		fmt.Println("Failed starting app")
		cancel()
		return
	}
	<-app.Done()
	_ = app.Stop(ctx)
}
