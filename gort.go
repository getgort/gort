package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/clockworksoul/gort/adapter"
	"github.com/clockworksoul/gort/adapter/slack"
	"github.com/clockworksoul/gort/config"
	"github.com/clockworksoul/gort/meta"
	"github.com/clockworksoul/gort/relay"
	"github.com/clockworksoul/gort/service"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var configfile string

var verboseCount int

var rootCmd = &cobra.Command{
	Use:   "gort",
	Short: "Bringing the power of the command line to chat",
	Long:  `Bringing the power of the command line to chat.`,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Immediately start the Gort server",
	Long:  `Immediately start the Gort server.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := startGort()
		if err != nil {
			log.Fatal(err.Error())
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Gort's version number",
	Long:  `All software has versions. This is Gort's.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Gort ChatOps Engine v%s\n", meta.GortVersion)
	},
}

func initializeCommands() {
	startCmd.Flags().StringVarP(
		&configfile,
		"config", "c", "config.yml", "The location of the config file to use")

	startCmd.Flags().CountVarP(
		&verboseCount,
		"verbose", "v", "Verbose mode (can be used multiple times)")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(versionCmd)
}

func initializeConfig(configfile string) error {
	err := config.Initialize(configfile)
	if err != nil {
		return err
	}

	config.BeginChangeCheck(3 * time.Second)

	return nil
}

func initializeLogger(verbose int) {
	log.SetFormatter(
		&log.TextFormatter{
			ForceColors:  true,
			PadLevelText: true,
		},
	)

	switch verbose {
	case 0:
		log.SetLevel(log.InfoLevel)
	case 1:
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.TraceLevel)
	}
}

func installAdapters() error {
	// TODO Add support for (and implementations of) other chat types.
	adapters := config.GetSlackProviders()

	if len(adapters) == 0 {
		return fmt.Errorf("no adapters configured")
	}

	for _, sp := range adapters {
		log.WithField("adapter.name", sp.Name).Info("Installing adapter")
		adapter.AddAdapter(slack.NewAdapter(sp))
	}

	return nil
}

func startGort() error {
	initializeLogger(verboseCount)

	log.WithField("version", meta.GortVersion).Infof("Starting Gort")

	// Load the Gort configuration.
	err := initializeConfig(configfile)
	if err != nil {
		return err
	}

	err = installAdapters()
	if err != nil {
		return err
	}

	// Start the Gort REST web service
	startServer(config.GetGortServerConfigs().APIAddress)

	// Listen for signals for graceful shutdown
	go catchSignals()

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
			log.WithError(aerr).Error("Error reported by adapter")
		}
	}
}

func catchSignals() {
	c := make(chan os.Signal, 1)

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	sig := <-c

	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on meta cancellation.
	log.WithField("signal", sig.String()).Info("Gracefully shutting down Gort")

	os.Exit(0)
}

func startServer(addr string) {
	// Build the service representation
	server := service.BuildRESTServer(addr)

	// Start watching the
	go func() {
		logs := server.Requests()
		for event := range logs {
			log.WithTime(event.Timestamp).
				WithField("addr", event.Addr).
				WithField("request", event.Request).
				WithField("size", event.Size).
				WithField("status", event.Status).
				WithField("user", event.UserID).
				Info("REST service event")
		}
	}()

	// Make the service listen.
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.WithError(err).Fatal("Fatal service error")
		}
	}()
}
