package main

import (
	"os"
	"os/signal"

	"github.com/elastos/Elastos.ELA.SPV/log"

	"github.com/elastos/Elastos.ELA.SPV_node/config"
	"github.com/elastos/Elastos.ELA.SPV_node/node"
	"github.com/elastos/Elastos.ELA.SPV_node/rpc"
)

func main() {
	log.Init()

	err := node.Init(config.Values().SeedList)
	if err != nil {
		log.Error("SPV node initialize failed, ", err)
		os.Exit(1)
	}

	// Handle interrupt signal
	stop := make(chan int, 1)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			log.Trace("SPV node shutting down...")
			node.Instance.Stop()
			stop <- 1
		}
	}()

	node.Instance.Start()
	go rpc.StartServer()

	<-stop
}
