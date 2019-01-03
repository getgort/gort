package relay

import (
	"log"

	"github.com/clockworksoul/cog2/config"
)

// ProviderInfo contains the basic information for a provider.
type ProviderInfo struct {
	Type string
	Name string
}

// NewProviderInfoFromConfig can create a ProviderInfo from a config.Provider instance.
func NewProviderInfoFromConfig(ap config.Provider) *ProviderInfo {
	return (&ProviderInfo{}).SetProviderInfo(ap)
}

// SetProviderInfo can set a ProviderInfo from a config.Provider instance.
func (p *ProviderInfo) SetProviderInfo(provider config.Provider) *ProviderInfo {
	switch ap := provider.(type) {
	case config.SlackProvider:
		p.Type = "slack"
		p.Name = ap.Name
	default:
		log.Panicf("Unsupported provider type: %T\n", ap)
	}

	return p
}
