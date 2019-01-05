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
func NewProviderInfoFromConfig(ap data.Provider) *ProviderInfo {
	return (&ProviderInfo{}).SetProviderInfo(ap)
}

// SetProviderInfo can set a ProviderInfo from a data.Provider instance.
func (p *ProviderInfo) SetProviderInfo(provider data.Provider) *ProviderInfo {
	switch ap := provider.(type) {
	case data.SlackProvider:
		p.Type = "slack"
		p.Name = ap.Name
	default:
		log.Errorf("Unsupported provider type: %T", ap)
	}

	return p
}
