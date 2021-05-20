package adapter

import (
	"github.com/clockworksoul/cog2/data"
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
		log.Errorf("[ProviderInfo.NewProviderInfoFromConfig] Unsupported provider type: %T", ap)
	}

	return p
}
