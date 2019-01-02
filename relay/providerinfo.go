package relay

import (
	"log"

	"github.com/clockworksoul/cog2/config"
)

type ProviderInfo struct {
	Type string
	Name string
}

func NewProviderInfoFromConfig(ap config.Provider) *ProviderInfo {
	return (&ProviderInfo{}).SetProviderInfo(ap)
}

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
