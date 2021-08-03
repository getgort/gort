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
	"github.com/getgort/gort/dataaccess/errs"
	"github.com/getgort/gort/dataaccess/memory"
	"github.com/getgort/gort/dataaccess/postgres"
	"github.com/getgort/gort/errors"
	"github.com/getgort/gort/telemetry"

	log "github.com/sirupsen/logrus"
)

var (
	// This will be non-nil if there was an error in initialization.
	// A successful initialization will reset it to nil.
	initializationError error

	// This is locked during the process of initializing.
	initializationMutex = sync.RWMutex{}
)

func init() {
	go monitorConfig(context.Background())
}

// Get provides an interface to the data access layer. If the last
// initialization attempt failed this will return an error.
func Get() (DataAccess, error) {
	if initializationError != nil {
		return nil, initializationError
	}

	initializationMutex.RLock()
	defer initializationMutex.RUnlock()

	return getCorrectDataAccess(), nil
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
func initializeDataAccess(ctx context.Context, cancel chan interface{}) {
	initializationMutex.Lock()

	go func() {
		defer initializationMutex.Unlock()

		delay := time.Second
		dataAccess := getCorrectDataAccess()
		err := dataAccess.Initialize(ctx)

		for err != nil {
			initializationError = errors.Wrap(errs.ErrDataAccessCantInitialize, err)

			log.WithError(err).Warn("Failed to connect to data source")
			telemetry.Errors().WithError(err).Commit(ctx)
			log.WithField("delay", delay).Info("Waiting to try again")

			select {
			case <-time.After(delay):
			case <-cancel:
				return
			case <-ctx.Done():
				initializationError = errors.Wrap(errs.ErrDataAccessCantInitialize, ctx.Err())
				log.WithError(ctx.Err()).Error("Could not initialize DAL")
				return
			}

			delay *= 2
			if delay > 10*time.Second {
				delay = 10 * time.Second
			}

			dataAccess = getCorrectDataAccess()
			err = dataAccess.Initialize(ctx)
		}

		initializationError = nil
		log.WithField("type", fmt.Sprintf("%T", dataAccess)).
			Info("Connection to data source established")
	}()
}

// monitorConfig monitors config.Updates(), and initializes the new DAL
// whenever a change is observed.
func monitorConfig(ctx context.Context) {
	configListener := config.Updates()
	cancel := make(chan interface{})

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
				cancel <- true
			}

			initializeDataAccess(ctx, cancel)
		}

		lastConfigState = cs
	}
}
