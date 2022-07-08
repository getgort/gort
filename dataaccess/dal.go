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

	"github.com/getgort/gort/bundles"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/rest"

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
	initializationMutex.RLock()
	defer initializationMutex.RUnlock()

	if initializationError != nil {
		return nil, initializationError
	}

	return getCorrectDataAccess(), nil
}

func getCorrectDataAccess() DataAccess {
	dbConfigs := config.GetDatabaseConfigs()

	if config.Undefined(dbConfigs) {
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

func bootstrapUserWithDefaults(user rest.User) (rest.User, error) {
	// If user doesn't have a defined email, we default to "gort@localhost".
	if user.Email == "" {
		user.Email = "gort@localhost"
	}

	// If user doesn't have a defined name, we default to "Gort Administrator".
	if user.FullName == "" {
		user.FullName = "Gort Administrator"
	}

	// The bootstrap user is _always_ named "admin".
	user.Username = "admin"

	// If user doesn't have a defined password, we kindly generate one.
	if user.Password == "" {
		password, err := data.GenerateRandomToken(32)
		if err != nil {
			return user, err
		}
		user.Password = password
	}

	return user, nil
}

func DoBootstrap(ctx context.Context, user rest.User) (rest.User, error) {
	const adminGroup = "admin"
	const adminRole = "admin"
	var adminPermissions = []string{
		"manage_commands",
		"manage_configs",
		"manage_groups",
		"manage_roles",
		"manage_users",
	}

	dataAccessLayer, err := Get()
	if err != nil {
		return user, err
	}

	// Set user defaults where necessary.
	user, err = bootstrapUserWithDefaults(user)
	if err != nil {
		return user, err
	}

	// Persist our shiny new user to the database.
	err = dataAccessLayer.UserCreate(ctx, user)
	if err != nil {
		return user, err
	}

	// Create admin group.
	err = dataAccessLayer.GroupCreate(ctx, rest.Group{Name: adminGroup})
	if err != nil {
		return user, err
	}

	// Add the admin user to the admin group.
	err = dataAccessLayer.GroupUserAdd(ctx, adminGroup, user.Username)
	if err != nil {
		return user, err
	}

	// Create an admin role
	err = dataAccessLayer.RoleCreate(ctx, adminRole)
	if err != nil {
		return user, err
	}

	// Add role to group
	err = dataAccessLayer.GroupRoleAdd(ctx, adminGroup, adminRole)
	if err != nil {
		return user, err
	}

	// Add the default permissions.
	for _, p := range adminPermissions {
		err = dataAccessLayer.RolePermissionAdd(ctx, adminRole, "gort", p)
		if err != nil {
			return user, err
		}
	}

	// Finally, add and enable the default bundle
	b, err := bundles.Default()
	if err != nil {
		return user, err
	}

	err = dataAccessLayer.BundleCreate(ctx, b)
	if err != nil {
		return user, err
	}

	err = dataAccessLayer.BundleEnable(ctx, b.Name, b.Version)
	if err != nil {
		return user, err
	}

	return user, nil
}
