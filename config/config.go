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

var (
	configfile       = "config.yml"
	md5sum           = []byte{}
	config           *data.CogConfig
	lastReloadWorked = true // Prevents spam
)

// BeginChangeCheck starts a routine that checks the underlying config for
// changes and reloads if one is found.
func BeginChangeCheck(frequency time.Duration) {
	ticker := time.NewTicker(frequency)

	go func() {
		for range ticker.C {
			err := ReloadConfiguration()
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
func Initialize(file string) error {
	configfile = file

	if _, err := os.Stat(configfile); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", configfile)
	}

	return ReloadConfiguration()
}

func ReloadConfiguration() error {
	sum, err := getMd5Sum(configfile)
	if err != nil {
		return fmt.Errorf("Failed hash file %s: %s", configfile, err.Error())
	}

	if !slicesAreEqual(sum, md5sum) {
		cp, err := loadConfiguration(configfile)
		if err != nil {
			return fmt.Errorf("Failed to load config %s: %s", configfile, err.Error())
		}

		md5sum = sum
		config = cp
		lastReloadWorked = true

		log.Infof("[ReloadConfiguration] Loaded configuration file %s", configfile)
	}

	return nil
}

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

// TODO Find some  way to automatically inject defaults if a field isn't specified.
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
