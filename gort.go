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

	"github.com/getgort/gort/adapter"
	"github.com/getgort/gort/adapter/discord"
	"github.com/getgort/gort/adapter/slack"
	"github.com/getgort/gort/config"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/relay"
	"github.com/getgort/gort/scheduler"
	"github.com/getgort/gort/service"
	"github.com/getgort/gort/telemetry"
	"github.com/getgort/gort/version"

	log "github.com/sirupsen/logrus"
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
	slackAdapters := config.GetSlackProviders()
	discordAdapters := config.GetDiscordProviders()

	if len(slackAdapters)+len(discordAdapters) == 0 {
		return fmt.Errorf("no adapters configured")
	}

	for _, sp := range slackAdapters {
		log.WithField("adapter.name", sp.Name).Info("Installing Slack adapter")
		adapter.AddAdapter(slack.NewAdapter(sp))
	}
	for _, sp := range discordAdapters {
		log.WithField("adapter.name", sp.Name).Info("Installing Discord adapter")
		ad, err := discord.NewAdapter(sp)
		if err != nil {
			return err
		}
		adapter.AddAdapter(ad)
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

	scheduledCommands := scheduler.StartScheduler()

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

		case request := <-scheduledCommands:
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

	var certFile, keyFile = config.TLSCertFile, config.TLSKeyFile
	var generated bool

	if certFile == "" || keyFile == "" {
		log.Warn("Generating TLS certificates, please consider getting real ones")

		cf, kf, err := service.GenerateTemporaryTLSKeys()
		if err != nil {
			log.WithError(err).Fatal("Failed to generate TLS certificates")
			return
		}

		certFile = cf
		keyFile = kf
		generated = true
	}

	// Start watching the request events
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

	// Make the service listen
	go func() {
		defer func() {
			if generated {
				os.Remove(certFile)
				os.Remove(keyFile)
			}
		}()

		if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
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
