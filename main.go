package main

import (
	"log"

	"github.com/clockworksoul/cog2/listeners"
)

func main() {
	log.Printf("Starting Cog2.%s\n", version)

	listeners.StartListening()

	<-make(chan struct{})
}
