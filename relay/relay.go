package relay

import (
	"github.com/clockworksoul/cog2/config"
)

func StartListening() {
	for _, sp := range config.GetSlackProviders() {
		listener := NewSlackRelay(sp)

		listener.Initialize()
		listener.Connect()

		go listener.Listen()
	}
}
