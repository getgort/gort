package main

import (
	"log"

	"github.com/clockworksoul/cog2/context"
	"github.com/clockworksoul/cog2/relays"
)

func main() {
	log.Printf("Starting Cog2 version %s\n", context.CogVersion)

	relays.StartListening()

	<-make(chan struct{})
}
