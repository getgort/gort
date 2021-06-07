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

package dataaccess

import (
	"context"
	"fmt"
	"time"

	"github.com/getgort/gort/config"
	"github.com/getgort/gort/dataaccess/postgres"
	"github.com/getgort/gort/telemetry"

	log "github.com/sirupsen/logrus"
)

const (
	// StateUninitialized is the default state of the data access
	// layer, before initialization begins.
	StateUninitialized State = iota

	// StateInitializing indicates that the data access layer is
	// initializing, either because it's new or because something has
	// changed (config updated, database disconnected, etc)
	StateInitializing

	// StateInitialized indicates the data access layer is fully
	// initialized and (presumably) functional.
	StateInitialized

	// StateError indicated that some kind of error is preventing the
	// data access layer from initializing correctly.
	StateError
)

var (
	badListenerEvents    = make(chan chan State) // Notified if there's an error emitting status
	configListener       <-chan config.State
	currentState         State
	dataAccessLayer      DataAccess
	stateChangeListeners []chan State
)

func init() {
	stateChangeListeners = make([]chan State, 0)

	go monitorConfig()
	go watchBadDALListenerEvents()
}

// State represents a possible state of the data access layer.
type State byte

func (s State) String() string {
	switch s {
	case StateUninitialized:
		return "StateUninitialized"
	case StateInitializing:
		return "StateInitializing"
	case StateInitialized:
		return "StateInitialized"
	case StateError:
		return "StateError"
	default:
		return "UNKNOWN STATE"
	}
}

// Get provides an interface to the data access layer. If the state is not
// 'initialized', this will return an error.
func Get() (DataAccess, error) {
	if currentState != StateInitialized {
		return nil, fmt.Errorf("data access layer not initialized")
	}

	return dataAccessLayer, nil
}

// CurrentState returns the current state of the data access layer.
func CurrentState() State {
	return currentState
}

// Updates returns a channel that blocks emits a signal whenever the data
// access layer state changes, typically on startup and whenever the
// configuration changes. Upon creation, it will emit the current state,
// so it never blocks.
func Updates() <-chan State {
	ch := make(chan State)
	stateChangeListeners = append(stateChangeListeners, ch)

	go func() {
		ch <- currentState
	}()

	return ch
}

// initializeDataAccess is called by monitorConfig to initialize the data
// access layer whenever the configuration is updated.
func initializeDataAccess() {
	configUpdates := config.Updates()

	// Ignore current status
	<-configUpdates

	go func() {
		var delay time.Duration = 1

		updateDALState(StateInitializing)

		for currentState != StateInitialized {
			dbConfigs := config.GetDatabaseConfigs()
			dataAccessLayer = postgres.NewPostgresDataAccess(dbConfigs) // hard-coded for now
			// dataAccessLayer = memory.NewInMemoryDataAccess()

			err := dataAccessLayer.Initialize()

			if err != nil {
				log.WithError(err).Warn("Failed to connect to data source")
				telemetry.Errors().WithError(err).Commit(context.TODO())
				log.WithField("delay", delay).Info("Waiting to try again")

				updateDALState(StateError)

				select {
				case <-time.After(delay * time.Second):
				case configStatus := <-configUpdates:
					// if this happens, then initializeDataAccess() was just called again.
					// Cancel this attempt.
					if configStatus == config.StateConfigInitialized {
						log.Debug("Starting over with new config")
						return
					}
				}

				delay *= 2

				if delay > 60 {
					delay = 60
				}
			} else {
				log.WithField("type", fmt.Sprintf("%T", dataAccessLayer)).
					Info("Connection to data source established")
				updateDALState(StateInitialized)
			}
		}
	}()
}

// monitorConfig monitors config.Updates(), and updates the data access
// singleton whenever a change is observed.
func monitorConfig() {
	configListener = config.Updates()

	// ConfigListener always emits the current state upon creation.
	lastConfigState := <-configListener

	for cs := range configListener {
		switch cs {
		case config.StateConfigUninitialized:
			fallthrough
		case config.StateConfigError:
			log.Info("Waiting for config to report initialized")
		case config.StateConfigInitialized:
			if lastConfigState != config.StateConfigUninitialized {
				log.Info("Configuration change detected: updating data access interface")
			}

			initializeDataAccess()
		}

		lastConfigState = cs
	}
}

// updateDALState updates the state and emits the new state to any listeners.
func updateDALState(newState State) {
	currentState = newState

	log.WithField("status", newState).Trace("Received status update")

	// Sadly, this needs to track and remove closed channels.
	for _, ch := range stateChangeListeners {
		updateDALStateTryEmit(ch, newState)
	}
}

// updateDALStateTryEmit will attempt to emit to a listener. If the channel is
// closed, it is removed from the listeners list. Blocking channels are ignored.
func updateDALStateTryEmit(ch chan State, newState State) {
	defer func() {
		if r := recover(); r != nil {
			// The channel was closed.
			badListenerEvents <- ch
		}
	}()

	select {
	case ch <- newState:
		// Everything is good
	default:
		// Channel is blocking. Ignore for now.
		// Eventually GC should close it and we can remove.
	}
}

// watchBadDALListenerEvents call be init to observe the badListenerEvents
// channel, and removes any bad channels that it hears about.
func watchBadDALListenerEvents() {
	badListenerEvents = make(chan chan State)

	log.Trace("Cleaning up closed channel")

	for chbad := range badListenerEvents {
		newChs := make([]chan State, 0)

		for _, ch := range stateChangeListeners {
			if chbad != ch {
				newChs = append(newChs, ch)
			}
		}

		stateChangeListeners = newChs
	}
}
