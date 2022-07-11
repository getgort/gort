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

package config

import (
	"crypto/md5"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/getgort/gort/data"
	gerrs "github.com/getgort/gort/errors"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const (
	EnvDatabasePassword = "GORT_DB_PASSWORD"
)

const (
	// StateConfigUninitialized is the default state of the config,
	// before initialization begins.
	StateConfigUninitialized State = iota

	// StateConfigInitialized indicates the config is fully initialized.
	StateConfigInitialized

	// StateConfigError indicates that the configuration file could not be
	// loaded and any calls to GetXConfig() will be invalid. This will only be
	// seen on the initial load attempt: if a config file changes and cannot be
	// loaded, the last good config will be used and the state will remain
	// 'initialized'
	StateConfigError
)

var (
	config      = &data.GortConfig{}
	configFile  string
	configMutex = sync.RWMutex{}
	md5sum      = []byte{}

	stateChangeListeners      = make([]chan State, 0)
	stateChangeListenersMutex = sync.Mutex{}

	currentState = StateConfigUninitialized
	stateMutex   = sync.RWMutex{}
)

var (
	// ErrConfigFileNotFound is returned by Initialize() if the specified
	// config file doesn't exist.
	ErrConfigFileNotFound = errors.New("config file doesn't exist")

	// ErrHashFailure can be returned by Initialize() or internal methods if
	// there's an error while generating a hash for the configuration file.
	ErrHashFailure = errors.New("failed to generate config file hash")

	// ErrConfigUnloadable can be returned by Initialize() or internal
	// methods if the config file exists but can't be loaded.
	ErrConfigUnloadable = errors.New("can't load config file")
)

// State represents a possible state of the config.
type State byte

func (s State) String() string {
	switch s {
	case StateConfigUninitialized:
		return "StateConfigUninitialized"
	case StateConfigInitialized:
		return "StateConfigInitialized"
	case StateConfigError:
		return "StateConfigError"
	default:
		return "UNKNOWN STATE"
	}
}

// CurrentState returns the current state of the config.
func CurrentState() State {
	stateMutex.RLock()
	defer stateMutex.RUnlock()

	return currentState
}

// GetDatabaseConfigs returns the data wrapper for the "database" config section.
func GetDatabaseConfigs() data.DatabaseConfigs {
	configMutex.RLock()
	defer configMutex.RUnlock()

	return config.DatabaseConfigs
}

// GetDiscordProviders returns the data wrapper for the "discord" config section.
func GetDiscordProviders() []data.DiscordProvider {
	configMutex.RLock()
	defer configMutex.RUnlock()

	return config.DiscordProviders
}

// GetDockerConfigs returns the data wrapper for the "docker" config section.
func GetDockerConfigs() data.DockerConfigs {
	configMutex.RLock()
	defer configMutex.RUnlock()

	return config.DockerConfigs
}

// GetDynamicConfigs returns the data wrapper for the "dynamic" config section.
func GetDynamicConfigs() data.DynamicConfigs {
	configMutex.RLock()
	defer configMutex.RUnlock()

	return config.DynamicConfigs
}

// GetGlobalConfigs returns the data wrapper for the "global" config section.
func GetGlobalConfigs() data.GlobalConfigs {
	configMutex.RLock()
	defer configMutex.RUnlock()

	return config.GlobalConfigs
}

// GetGortServerConfigs returns the data wrapper for the "gort" config section.
func GetGortServerConfigs() data.GortServerConfigs {
	configMutex.RLock()
	defer configMutex.RUnlock()

	return config.GortServerConfigs
}

// GetJaegerConfigs returns the data wrapper for the "jaeger" config section.
func GetJaegerConfigs() data.JaegerConfigs {
	configMutex.RLock()
	defer configMutex.RUnlock()

	return config.JaegerConfigs
}

// GetKubernetesConfigs returns the data wrapper for the "jaeger" config section.
func GetKubernetesConfigs() data.KubernetesConfigs {
	configMutex.RLock()
	defer configMutex.RUnlock()

	return config.KubernetesConfigs
}

// GetSlackProviders returns the data wrapper for the "slack" config section.
func GetSlackProviders() []data.SlackProvider {
	configMutex.RLock()
	defer configMutex.RUnlock()

	return config.SlackProviders
}

// GetTemplates returns the deployment-scoped template overrides.
func GetTemplates() data.Templates {
	configMutex.RLock()
	defer configMutex.RUnlock()

	return config.Templates
}

// Initialize is called by main() to trigger creation of the config singleton.
// It can be called multiple times, if you're into that kind of thing. If
// successful, this will emit a StateConfigInitialized to any update listeners.
func Initialize(file string) error {
	configFile = file

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		updateConfigState(StateConfigError)
		return gerrs.Wrap(ErrConfigFileNotFound, err)
	}

	return Reload()
}

// Undefined is a helper method that is used to determine whether config
// sections are present.
func Undefined(c interface{}) bool {
	if c == nil {
		return true
	}

	return reflect.ValueOf(c).IsZero()
}

// Reload is called by Initialize() to determine whether the config file has
// changed (or is new) and reload if it has.
func Reload() error {
	configMutex.Lock()
	defer configMutex.Unlock()

	sum, err := getMd5Sum(configFile)
	if err != nil {
		log.WithField("file", configFile).WithError(err).Error(ErrHashFailure.Error())

		return gerrs.Wrap(ErrHashFailure, err)
	}

	if !slicesAreEqual(sum, md5sum) {
		cp, err := load(configFile)
		if err != nil {
			// If we're already initialized, keep the original config.
			// If not, set the state to 'error'.
			if CurrentState() == StateConfigUninitialized {
				updateConfigState(StateConfigError)
			}

			log.WithField("file", configFile).WithError(err).Error(ErrConfigUnloadable.Error())

			return gerrs.Wrap(ErrConfigUnloadable, err)
		}

		md5sum = sum
		config = cp

		setLogFormatter()

		// Properly load the database configs.
		standardizeDatabaseConfig(&cp.DatabaseConfigs)

		updateConfigState(StateConfigInitialized)

		log.WithField("file", configFile).Info("Loaded configuration file")
	}

	return nil
}

// Updates returns a channel that emits a message whenever the underlying
// configuration is updated. Upon creation, it will emit the current state,
// so it never blocks.
func Updates() <-chan State {
	stateChangeListenersMutex.Lock()
	defer stateChangeListenersMutex.Unlock()

	ch := make(chan State, 1)
	stateChangeListeners = append(stateChangeListeners, ch)
	ch <- CurrentState()

	return ch
}

// getMd5Sum is used to determine when the underlying config file is modified.
func getMd5Sum(file string) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return []byte{}, gerrs.Wrap(gerrs.ErrIO, err)
	}
	defer f.Close()

	hasher := md5.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return []byte{}, gerrs.Wrap(gerrs.ErrIO, err)
	}

	hashBytes := hasher.Sum(nil)

	return hashBytes, nil
}

// load creates a new GortConfig from a file. It's usually called
// by Reload() to execute the actual steps of loading the
// configuration.
func load(file string) (*data.GortConfig, error) {
	// Read file as a byte slice
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, gerrs.Wrap(gerrs.ErrIO, err)
	}

	var config data.GortConfig

	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		return nil, gerrs.Wrap(gerrs.ErrUnmarshal, err)
	}

	return &config, nil
}

func setLogFormatter() {
	dev := config.GortServerConfigs.DevelopmentMode

	if dev {
		log.SetFormatter(
			&log.TextFormatter{
				ForceColors:  true,
				PadLevelText: true,
			},
		)
	} else {
		log.SetFormatter(&log.JSONFormatter{})
	}

	log.WithField("development", dev).Debug("Log formatter defined")
}

// slicesAreEqual compares two hashcode []bytes and returns true if they're
// identical.
func slicesAreEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

func standardizeDatabaseConfig(dbc *data.DatabaseConfigs) {
	if dbc.Password == "" {
		log.Debug("Config database password empty; using envvar ", EnvDatabasePassword)

		if dbc.Password = os.Getenv(EnvDatabasePassword); dbc.Password == "" {
			log.Debug("Config database password cannot be found")
		}
	}
}

// updateConfigState updates the state and emits the new state to any listeners.
func updateConfigState(newState State) {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	currentState = newState

	log.WithField("state", newState).Trace("Config state update")

	for _, ch := range stateChangeListeners {
		timer := time.NewTimer(100 * time.Millisecond)

		select {
		case ch <- newState:
		case <-timer.C:
		}
	}
}
