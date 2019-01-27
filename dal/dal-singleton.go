package dal

import (
	"fmt"
	"time"

	"github.com/clockworksoul/cog2/config"
	"github.com/clockworksoul/cog2/dal/memory"
	"github.com/clockworksoul/cog2/dal/postgres"

	log "github.com/sirupsen/logrus"
)

var (
	dataAccessLayerInitializing bool
	dataAccessLayerInitialized  bool
	dataAccessLayer             DataAccess
	initializationListeners     []chan struct{}
)

func init() {
	initializationListeners = make([]chan struct{}, 0)
}

// DataAccessInterface provides an interface to the data access layer.
func DataAccessInterface() (DataAccess, error) {
	initializeDataAccess()

	if !dataAccessLayerInitialized {
		return memory.NewInMemoryDataAccess(), fmt.Errorf("data access not initialized")
	}

	return dataAccessLayer, nil
}

// ListenForInitialization returns a channel that blocks until the DAL
// is initialized.
func ListenForInitialization() <-chan struct{} {
	initializeDataAccess()

	ch := make(chan struct{}, 1)

	if dataAccessLayerInitialized {
		go func() {
			ch <- struct{}{}
		}()
	} else {
		initializationListeners = append(initializationListeners, ch)
	}

	return ch
}

// initializeDataAccess will initialize the data access layer, if it isn't
// already initialized.
func initializeDataAccess() {
	if dataAccessLayerInitializing {
		return
	}

	dataAccessLayerInitializing = true

	go func() {
		var delay time.Duration = 1

		for !dataAccessLayerInitialized {
			dbConfigs := config.GetDatabaseConfigs()
			dataAccessLayer = postgres.NewPostgresDataAccess(dbConfigs) // hard-coded for now
			err := dataAccessLayer.Initialize()

			if err != nil {
				log.Warn("[initializeDataAccess] Failed to connect to data source: ", err.Error())
				log.Infof("[initializeDataAccess] Waiting %d seconds to try again", delay)

				<-time.After(delay * time.Second)

				delay *= 2

				if delay > 60 {
					delay = 60
				}
			} else {
				dataAccessLayerInitialized = true
			}
		}

		for _, ch := range initializationListeners {
			ch <- struct{}{}
		}

		log.Info("[initializeDataAccess] Connection to data source established")
	}()
}
