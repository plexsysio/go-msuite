package main

import (
	"context"
	"fmt"
	"github.com/plexsysio/go-msuite/lib"
	logger "github.com/ipfs/go-log/v2"
)

func main() {
	logger.SetLogLevel("*", "Debug")

	ctx, cancel := context.WithCancel(context.Background())

	app, err := msuite.New(
		msuite.WithHTTP(10000),
		msuite.WithJWT("dummysecret"),
		msuite.WithP2PPort(10001),
		msuite.WithGRPCTCPListener(10002),
		msuite.WithServiceACL(nil),
		msuite.WithPrometheus(true),
	)
	fmt.Println("Starting")
	err = app.Start(ctx)
	if err != nil {
		fmt.Println("Failed starting app", err.Error())
		cancel()
		return
	}
	<-app.Done()
	_ = app.Stop(ctx)
}
