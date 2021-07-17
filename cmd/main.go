package main

import (
	"context"
	"fmt"
	logger "github.com/ipfs/go-log/v2"
	"github.com/plexsysio/go-msuite/lib"
)

func main() {
	_ = logger.SetLogLevel("*", "Debug")

	ctx := context.Background()

	app, err := msuite.New(
		msuite.WithHTTP(10000),
		msuite.WithJWT("dummysecret"),
		msuite.WithP2PPort(10001),
		msuite.WithGRPCTCPListener(10002),
		msuite.WithServiceACL(nil),
		msuite.WithPrometheus(true),
	)
	if err != nil {
		fmt.Printf("failed creating go-msuite node %s\n", err.Error())
		return
	}
	fmt.Println("Starting")
	err = app.Start(ctx)
	if err != nil {
		fmt.Println("failed starting app", err.Error())
		return
	}
	<-app.Done()
	_ = app.Stop(ctx)
}
