package main

import (
	"os"
	"os/signal"

	"github.com/elastos/Elastos.ELA.SPV/log"

	"github.com/elastos/Elastos.ELA.SPV.Node/config"
	"github.com/elastos/Elastos.ELA.SPV.Node/node"
	"github.com/elastos/Elastos.ELA.SPV.Node/rpc"
)

func main() {
	log.Init(config.Values().PrintLevel)

	spvNode, err := node.NewSpvNode(config.Values().SeedList)
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
			spvNode.Stop()
			stop <- 1
		}
	}()

	go rpc.StartServer(spvNode)
	spvNode.Start()

	<-stop
}
