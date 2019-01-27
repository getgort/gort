package config

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/clockworksoul/cog2/data"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

const (
	// StateConfigUninitialized is the default state of the config,
	// before initializtion begins.
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
	badListenerEvents    = make(chan chan State) // Notified if there's an error emitting status
	config               *data.CogConfig
	configFile           string
	currentState         = StateConfigUninitialized
	lastReloadWorked     = true // Used to keep prevent log spam
	md5sum               = []byte{}
	stateChangeListeners = make([]chan State, 0)
)

func init() {
	go watchBadConfigListenerEvents()
}

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

// BeginChangeCheck starts a routine that checks the underlying config for
// changes and reloads if one is found.
func BeginChangeCheck(frequency time.Duration) {
	ticker := time.NewTicker(frequency)

	go func() {
		for range ticker.C {
			err := reloadConfiguration()
			if err != nil {
				if lastReloadWorked {
					lastReloadWorked = false
					log.Errorf("[BeginChangeCheck] %s", err.Error())
				}
			}
		}
	}()
}

// GetBundleConfigs returns the data wrapper for the "bundles" config section.
func GetBundleConfigs() []data.Bundle {
	return config.BundleConfigs
}

// GetDatabaseConfigs returns the data wrapper for the "database" config section.
func GetDatabaseConfigs() data.DatabaseConfigs {
	return config.DatabaseConfigs
}

// GetDockerConfigs returns the data wrapper for the "docker" config section.
func GetDockerConfigs() data.DockerConfigs {
	return config.DockerConfigs
}

// GetCogServerConfigs returns the data wrapper for the "cog" config section.
func GetCogServerConfigs() data.CogServerConfigs {
	return config.CogServerConfigs
}

// GetGlobalConfigs returns the data wrapper for the "global" config section.
func GetGlobalConfigs() data.GlobalConfigs {
	return config.GlobalConfigs
}

// GetSlackProviders returns the data wrapper for the "slack" config section.
func GetSlackProviders() []data.SlackProvider {
	return config.SlackProviders
}

// Initialize is called by main() to trigger creation of the config singleton.
// It can be called multiple times, if you're into that kind of thing. If
// succesful, this will emit a StateConfigInitialized to any update listeners.
func Initialize(file string) error {
	configFile = file

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		updateConfigState(StateConfigError)
		return fmt.Errorf("file %s does not exist", configFile)
	}

	return reloadConfiguration()
}

// CurrentState returns the current state of the config.
func CurrentState() State {
	return currentState
}

// Updates returns a channel that emits a message whenever the underlying
// configuration is updated. Upon creation, it will emit the current state,
// so it never blocks.
func Updates() <-chan State {
	ch := make(chan State)
	stateChangeListeners = append(stateChangeListeners, ch)

	go func() {
		ch <- currentState
	}()

	return ch
}

// getMd5Sum is used to determine when the underlying config file is modified.
func getMd5Sum(file string) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return []byte{}, err
	}
	defer f.Close()

	hasher := md5.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return []byte{}, err
	}

	hashBytes := hasher.Sum(nil)

	return hashBytes, nil
}

// loadConfiguration is called by reloadConfiguration() to execute the actual
// steps of loading the configuration.
func loadConfiguration(file string) (*data.CogConfig, error) {
	// Read file as a byte slice
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var config data.CogConfig

	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

//  reloadConfiguration is called by both BeginChangeCheck() and Initialize()
// to determine whether the config file has changed (or is new) and reload if
// it has.
func reloadConfiguration() error {
	sum, err := getMd5Sum(configFile)
	if err != nil {
		return fmt.Errorf("Failed hash file %s: %s", configFile, err.Error())
	}

	if !slicesAreEqual(sum, md5sum) {
		cp, err := loadConfiguration(configFile)
		if err != nil {
			// If we're already initialized, keep the original config.
			// If not, set the state to 'error'.
			if currentState == StateConfigUninitialized {
				updateConfigState(StateConfigError)
			}

			return fmt.Errorf("Failed to load config %s: %s", configFile, err.Error())
		}

		md5sum = sum
		config = cp
		lastReloadWorked = true

		updateConfigState(StateConfigInitialized)

		log.Infof("[reloadConfiguration] Loaded configuration file %s", configFile)
	}

	return nil
}

// slicesAreEqual compares two []bytes and returns true if they're identical.
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

// updateConfigState updates the state and emits the new state to any listeners.
func updateConfigState(newState State) {
	currentState = newState

	log.Tracef("[updateConfigState] Received status update")

	// Sadly, this needs to track and remove closed channels.
	for _, ch := range stateChangeListeners {
		updateConfigStateTryEmit(ch, newState)
	}
}

// updateConfigStateTryEmit will attempt to emit to a listener. If the channel is
// closed or the send times out, it is removed from the listeners list.
func updateConfigStateTryEmit(ch chan State, newState State) {
	defer func() {
		if r := recover(); r != nil {
			badListenerEvents <- ch
		}
	}()

	var ok = false // If not set to true, updateConfigStateEmitTimeout() closes the channel
	go updateConfigStateEmitTimeout(ch, &ok, time.Second)
	ch <- newState
	ok = true
}

// updateConfigStateEmitTimeout is called by updateConfigStateTryEmit. If ok isn't true
// after duration, ch is closed.
func updateConfigStateEmitTimeout(ch chan State, ok *bool, duration time.Duration) {
	defer func() {
		if r := recover(); r != nil {
			badListenerEvents <- ch
		}
	}()

	<-time.After(duration)

	if !*ok {
		close(ch)
	}
}

// watchBadConfigListenerEvents call be init to observe the badListenerEvents
// channel, and removes any bad channels that it hears about.
func watchBadConfigListenerEvents() {
	badListenerEvents = make(chan chan State)

	log.Tracef("[watchBadConfigListenerEvents] Cleaning up closed channel")

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
