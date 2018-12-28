package main

import (
	"log"

	"github.com/clockworksoul/cog2/context"
	"github.com/clockworksoul/cog2/relay"
)

func main() {
	log.Printf("Starting Cog2 version %s\n", context.CogVersion)

	relay.StartListening()

	<-make(chan struct{})
}
