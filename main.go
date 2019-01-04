package main

import (
	"log"
	"time"

	"github.com/clockworksoul/cog2/config"
	"github.com/clockworksoul/cog2/context"
	"github.com/clockworksoul/cog2/relay"
)

func main() {
	log.Printf("Starting Cog2 version %s\n", context.CogVersion)

	err := config.Initialize("config.yml")
	if err != nil {
		panic(err.Error())
	}
	config.BeginChangeCheck(3 * time.Second)

	relay.StartListening()

	<-make(chan struct{})
}
