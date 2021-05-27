package adapter

import (
	"fmt"

	"github.com/clockworksoul/gort/data"
	log "github.com/sirupsen/logrus"
)

// ProviderInfo contains the basic information for a chat provider.
type ProviderInfo struct {
	Type string
	Name string
}

// NewProviderInfoFromConfig can create a ProviderInfo from a data.Provider instance.
func NewProviderInfoFromConfig(provider data.Provider) *ProviderInfo {
	p := &ProviderInfo{}

	switch ap := provider.(type) {
	case data.SlackProvider:
		p.Type = "slack"
		p.Name = ap.Name
	default:
		log.WithField("type", fmt.Sprintf("%T", ap)).
			Errorf("Unsupported provider type")
	}

	return p
}
