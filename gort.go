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
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/getgort/gort/adapter"
	"github.com/getgort/gort/adapter/slack"
	"github.com/getgort/gort/config"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/relay"
	"github.com/getgort/gort/service"
	"github.com/getgort/gort/telemetry"
	"github.com/getgort/gort/version"
)

func initializeConfig(configFile string) error {
	err := config.Initialize(configFile)
	if err != nil {
		return err
	}

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

func startGort(ctx context.Context, configFile string, verboseCount int) error {
	setLoggerVerbosity(verboseCount)

	go catchSignals()

	// Load the Gort configuration.
	err := initializeConfig(configFile)
	if err != nil {
		return err
	}

	err = telemetry.CreateAndRegisterExporters()
	if err != nil {
		return err
	}

	log.WithField("version", version.Version).Infof("Starting Gort")

	err = installAdapters()
	if err != nil {
		return err
	}

	// Start the Gort REST web service
	startServer(ctx, config.GetGortServerConfigs())

	// Tells the chat provider adapters (as defined in the config) to connect.
	// Returns channels to get user command requests and adapter errors out.
	requestsFrom, responsesTo, adapterErrorsFrom := adapter.StartListening(ctx)

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
			telemetry.Errors().WithError(aerr).Commit(ctx)
			log.WithError(aerr).Error("Error reported by adapter")
		}
	}
}

func catchSignals() {
	c := make(chan os.Signal, 1)

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C).
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT)

	for sig := range c {
		switch sig {
		case syscall.SIGINT:
			log.WithField("signal", sig.String()).
				Info("SIGINT: Gracefully shutting down Gort")
			os.Exit(0)
		case syscall.SIGHUP:
			log.WithField("signal", sig.String()).
				Info("SIGHUP: Reloading configuration")
			config.Reload()
		}
	}
}

func startServer(ctx context.Context, config data.GortServerConfigs) {
	// Build the service representation
	server := service.BuildRESTServer(ctx, config.APIAddress)

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
			telemetry.Errors().WithError(err).Commit(ctx)
			log.WithError(err).Fatal("Fatal service error")
		}
	}()
}

func main() {
	err := GetRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
