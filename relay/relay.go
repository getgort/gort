package relay

import (
	"strings"

	"github.com/clockworksoul/cog2/config"
	"github.com/clockworksoul/cog2/worker"
)

type Relay interface {
	Listen()
	SendMessage(channel string, message string)
}

func FindImage(commandString string) (string, error) {
	return "clockworksoul/echotest", nil
}

func TokenizeParameters(commandString string) []string {
	return strings.Split(commandString, " ")
}

func SpawnWorker(imageName string, cmdParameters []string) (<-chan string, error) {
	worker, err := worker.NewWorker(imageName, cmdParameters...)
	if err != nil {
		return nil, err
	}

	return worker.Start()
}

func StartListening() {
	for _, sp := range config.GetSlackProviders() {
		listener := NewSlackRelay(sp)

		go listener.Listen()
	}
}
