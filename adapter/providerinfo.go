/*
 * Copyright 2021 The Gort Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
