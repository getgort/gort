package relay

import (
	"github.com/clockworksoul/cog2/config"
)

type Relay interface {
	Listen()
}

func StartListening() {
	for _, sp := range config.GetSlackProviders() {
		listener := NewSlackRelay(sp)

		go listener.Listen()
	}
}
