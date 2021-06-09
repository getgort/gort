/*
 * Copyright 2021 The Gort Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/getgort/gort/adapter"
	"github.com/getgort/gort/adapter/slack"
	"github.com/getgort/gort/config"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/relay"
	"github.com/getgort/gort/service"
	"github.com/getgort/gort/telemetry"
	"github.com/getgort/gort/version"
)

var (
	cmdRootConfigfile string

	cmdRootVerboseCount int

	cmdRoot = &cobra.Command{
		Use:   "gort",
		Short: "Bringing the power of the command line to chat",
		Long:  `Bringing the power of the command line to chat.`,
	}
)

var (
	cmdStart = &cobra.Command{
		Use:   "start",
		Short: "Immediately start the Gort server",
		Long:  `Immediately start the Gort server.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := startGort()
			if err != nil {
				log.WithError(err).Fatal("Fatal service error")
			}
		},
	}
)

var (
	cmdVersionShort bool

	cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "Print Gort's version number",
		Long:  `All software has versions. This is Gort's.`,
		Run: func(cmd *cobra.Command, args []string) {
			if cmdVersionShort {
				fmt.Println(version.Version)
				return
			}

			fmt.Printf("Gort ChatOps Engine v%s\n", version.Version)
		},
	}
)

func initializeCommands() {
	cmdStart.Flags().StringVarP(
		&cmdRootConfigfile,
		"config", "c", "config.yml", "The location of the config file to use")

	cmdStart.Flags().CountVarP(
		&cmdRootVerboseCount,
		"verbose", "v", "Verbose mode (can be used multiple times)")

	cmdRoot.AddCommand(cmdStart)

	cmdVersion.Flags().BoolVarP(
		&cmdVersionShort,
		"short", "s", false, "Print only the version number")

	cmdRoot.AddCommand(cmdVersion)
}

func initializeConfig(cmdRootConfigfile string) error {
	err := config.Initialize(cmdRootConfigfile)
	if err != nil {
		return err
	}

	config.BeginChangeCheck(3 * time.Second)

	return nil
}

func setLoggerVerbosity(verbose int) {
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
	setLoggerVerbosity(cmdRootVerboseCount)

	// Load the Gort configuration.
	err := initializeConfig(cmdRootConfigfile)
	if err != nil {
		return err
	}

	log.WithField("version", version.Version).Infof("Starting Gort")

	err = installAdapters()
	if err != nil {
		return err
	}

	// Start the Gort REST web service
	startServer(config.GetGortServerConfigs())

	// Listen for signals for graceful shutdown
	go catchSignals()

	// Tells the chat provider adapters (ad defined in the config) to connect.
	// Returns channels to get user command requests and adapter errors out.
	requestsFrom, responsesTo, adapterErrorsFrom := adapter.StartListening()

	// Starts the relay (currently just a local goroutine).
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
			telemetry.Errors().WithError(aerr).Commit(context.TODO())
			log.WithError(aerr).Error("Error reported by adapter")
		}
	}
}

func catchSignals() {
	c := make(chan os.Signal, 1)

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C).
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	sig := <-c

	log.WithField("signal", sig.String()).
		Info("Gracefully shutting down Gort")

	os.Exit(0)
}

func startServer(config data.GortServerConfigs) {
	// Build the service representation
	server := service.BuildRESTServer(config.APIAddress)

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
		var err error
		if config.TLSCertFile != "" && config.TLSKeyFile != "" {
			err = server.ListenAndServeTLS(config.TLSCertFile, config.TLSKeyFile)
		} else {
			log.Warn("Using http for API connections, please consider using https")
			err = server.ListenAndServe()
		}
		if err != nil {
			telemetry.Errors().WithError(err).Commit(context.TODO())
			log.WithError(err).Fatal("Fatal service error")
		}
	}()
}
