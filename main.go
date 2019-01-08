package main

import (
	"time"

	"github.com/clockworksoul/cog2/adapter"
	"github.com/clockworksoul/cog2/config"
	"github.com/clockworksoul/cog2/context"
	"github.com/clockworksoul/cog2/relay"
	"github.com/clockworksoul/cog2/service"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}

func initializeConfig(configfile string) error {
	err := config.Initialize(configfile)
	if err != nil {
		return err
	}

	config.BeginChangeCheck(3 * time.Second)

	return nil
}

func startRESTServer(addr string) {
	// Build the service representation
	server := service.BuildRESTServer(addr)

	// Start watching the
	go func() {
		logs := server.Requests()
		for logevent := range logs {
			log.Info(logevent)
		}
	}()

	// Make the service listen.
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Errorf("[main] %s", err.Error())
		}
	}()
}

func main() {
	log.Infof("[main] Starting Cog2 version %s", context.CogVersion)

	// Load the Cog configuration.
	// TODO Allow the user to specify a config document.
	err := initializeConfig("config.yml")
	if err != nil {
		log.Panicf("[main] %s", err.Error())
	}

	// Start the Cog REST web service
	startRESTServer(":8080")

	// Tells the chat provider adapters (ad defined in the config) to connect.
	// Returns channels to get user command requests and adapter errors out.
	requestsFrom, responsesTo, adapterErrorsFrom := adapter.StartListening()

	// Starts the relay (currently just a local gofunc).
	// Returns channels to send user command request in and get command
	// responses out.
	requestsTo, responsesFrom := relay.StartListening()

	for {
		select {
		// A user command request is received from a chat provider adapter.
		// Forward it to the relay.
		case request := <-requestsFrom:
			requestsTo <- request

		// A user command response is received from the relay.
		// Send it back to the adapter manager.
		case response := <-responsesFrom:
			responsesTo <- response

		// An adapter is reporting an error.
		case aerr := <-adapterErrorsFrom:
			log.Errorf("[main] %s", aerr.Error())
		}
	}
}
