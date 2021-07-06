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
	"sync"
	"time"

	"github.com/getgort/gort/config"
	"github.com/getgort/gort/dataaccess/memory"
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
	// badListenerEvents    = make(chan chan State) // Notified if there's an error emitting status
	configListener = config.Updates()
	dataAccess     DataAccess

	currentState State
	stateMutex   = sync.Mutex{}

	initializationMutex = sync.Mutex{}

	stateChangeListeners      []chan State
	stateChangeListenersMutex = sync.RWMutex{}

	configUpdates = config.Updates()
)

func init() {
	stateChangeListeners = make([]chan State, 0)

	go monitorConfig(context.Background())
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
	if CurrentState() != StateInitialized {
		return nil, fmt.Errorf("data access layer not initialized")
	}

	return dataAccess, nil
}

// CurrentState returns the current state of the data access layer.
func CurrentState() State {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	return currentState
}

// Updates returns a channel that blocks emits a signal whenever the data
// access layer state changes, typically on startup and whenever the
// configuration changes. Upon creation, it will emit the current state,
// so it never blocks.
func Updates() <-chan State {
	stateChangeListenersMutex.Lock()
	defer stateChangeListenersMutex.Unlock()

	ch := make(chan State, 1)
	stateChangeListeners = append(stateChangeListeners, ch)
	ch <- CurrentState()

	return ch
}

func getCorrectDataAccess() DataAccess {
	dbConfigs := config.GetDatabaseConfigs()

	if config.IsUndefined(dbConfigs) {
		return memory.NewInMemoryDataAccess()
	}

	return postgres.NewPostgresDataAccess(dbConfigs)
}

// initializeDataAccess is called by monitorConfig to initialize the data
// access layer whenever the configuration is updated.
func initializeDataAccess(ctx context.Context) {
	initializationMutex.Lock()

	// Ignore current status
	<-configUpdates

	go func() {
		defer initializationMutex.Unlock()

		delay := time.Second

		updateDALState(StateInitializing)

		for CurrentState() != StateInitialized {
			dataAccess = getCorrectDataAccess()
			err := dataAccess.Initialize(ctx)

			if err != nil {
				log.WithError(err).Warn("Failed to connect to data source")
				telemetry.Errors().WithError(err).Commit(ctx)
				log.WithField("delay", delay).Info("Waiting to try again")

				updateDALState(StateError)

				select {
				case <-time.After(delay):
				case configStatus := <-configUpdates:
					// if this happens, then initializeDataAccess() was just called again.
					// Cancel this attempt.
					if configStatus == config.StateConfigInitialized {
						log.Debug("Starting over with new config")
						return
					}
				case <-ctx.Done():
					log.WithError(ctx.Err()).Error("Could not initialize DAL")
					return
				}

				delay *= 2
				if delay > 10*time.Second {
					delay = 10 * time.Second
				}
			} else {
				log.WithField("type", fmt.Sprintf("%T", dataAccess)).
					Info("Connection to data source established")
				updateDALState(StateInitialized)
			}
		}
	}()
}

// monitorConfig monitors config.Updates(), and updates the data access
// singleton whenever a change is observed.
func monitorConfig(ctx context.Context) {
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

			initializeDataAccess(context.Background())
		}

		lastConfigState = cs
	}
}

// updateDALState updates the state and emits the new state to any listeners.
func updateDALState(newState State) {
	stateMutex.Lock()
	currentState = newState
	stateMutex.Unlock()

	log.WithField("state", newState).Trace("Config status update")

	stateChangeListenersMutex.RLock()
	defer stateChangeListenersMutex.RUnlock()

	for _, ch := range stateChangeListeners {
		timer := time.NewTimer(100 * time.Millisecond)

		select {
		case ch <- newState:
		case <-timer.C:
		}
	}
}
