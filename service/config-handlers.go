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

package service

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess/errs"
	"github.com/getgort/gort/dynamic"
	gerrs "github.com/getgort/gort/errors"
)

const (
	ParameterConfigurationLayer  = "layer"
	ParameterConfigurationBundle = "bundle"
	ParameterConfigurationOwner  = "owner"
	ParameterConfigurationKey    = "key"
)

func getDynamicConfigParameters(params map[string]string) (layer data.ConfigurationLayer, bundle string, owner string, key string, err error) {
	layer = data.ConfigurationLayer(params[ParameterConfigurationLayer])
	bundle = params[ParameterConfigurationBundle]
	owner = params[ParameterConfigurationOwner]
	key = params[ParameterConfigurationKey]
	err = layer.Validate()

	if layer == data.LayerBundle {
		owner = ""
	}

	return
}

// handleDeleteDynamicConfig handles "DELETE /v2/configs/{bundle}/{layer}/{owner}/{key}"
func handleDeleteDynamicConfig(w http.ResponseWriter, r *http.Request) {
	dc, err := dynamic.Get()
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	layer, bundle, owner, key, err := getDynamicConfigParameters(mux.Vars(r))
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	if layer == data.LayerUser {
		user, err := getUserByRequest(r)
		if err != nil {
			respondAndLogError(r.Context(), w, err)
			return
		}

		if owner == "" {
			owner = user.Username
		} else if owner != user.Username {
			respondAndLogError(r.Context(), w, ErrUnauthorized)
			return
		}
	}

	err = dc.Delete(r.Context(), layer, bundle, owner, key)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
}

// handleGetDynamicConfigs handles "GET /v2/configs"
func handleGetDynamicConfigs(w http.ResponseWriter, r *http.Request) {
	dc, err := dynamic.Get()
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	params := mux.Vars(r)
	p := func(key string) string {
		if val := params[key]; val == "*" {
			return ""
		} else {
			return val
		}
	}

	layer := data.ConfigurationLayer(p(ParameterConfigurationLayer))
	bundle := p(ParameterConfigurationBundle)
	owner := p(ParameterConfigurationOwner)
	key := p(ParameterConfigurationKey)

	configs, err := dc.List(r.Context(), layer, bundle, owner, key)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	} else if len(configs) == 0 {
		http.Error(w, "No dynamic configurations found", http.StatusNoContent)
		return
	}

	// Filter to ensure only the accessible configs are returned
	user, err := getUserByRequest(r)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	var filtered []data.DynamicConfiguration

	for _, c := range configs {
		if c.Layer == data.LayerUser && c.Owner != user.Username {
			continue
		}

		if c.Secret {
			c.Value = ""
		}

		filtered = append(filtered, c)
	}

	json.NewEncoder(w).Encode(filtered)
}

// handlePutDynamicConfiguration handles "PUT /v2/configs/{bundle}/{layer}/{owner}/{key}"
func handlePutDynamicConfiguration(w http.ResponseWriter, r *http.Request) {
	dc, err := dynamic.Get()
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	var c data.DynamicConfiguration
	err = json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		respondAndLogError(r.Context(), w, gerrs.ErrUnmarshal)
		return
	}

	layer, bundle, owner, key, err := getDynamicConfigParameters(mux.Vars(r))
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	if layer == data.LayerUser {
		user, err := getUserByRequest(r)
		if err != nil {
			respondAndLogError(r.Context(), w, err)
			return
		}

		if owner == "" {
			owner = user.Username
		} else if owner != user.Username {
			respondAndLogError(r.Context(), w, ErrUnauthorized)
			return
		}
	}

	if strings.HasPrefix(key, "GORT_") {
		respondAndLogError(r.Context(), w, errs.ErrConfigIllegal)
		return
	}

	c.Layer = layer
	c.Bundle = bundle
	c.Owner = owner
	c.Key = key

	err = dc.Set(r.Context(), c)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
}

func addConfigMethodsToRouter(router *mux.Router) {
	router.Handle("/v2/configs/{bundle}", otelhttp.NewHandler(authCommand(handleGetDynamicConfigs, "config", "get"), "handleGetConfigs")).Methods("GET")
	router.Handle("/v2/configs/{bundle}/{layer}", otelhttp.NewHandler(authCommand(handleGetDynamicConfigs, "config", "get"), "handleGetConfigs")).Methods("GET")
	router.Handle("/v2/configs/{bundle}/{layer}/{owner}", otelhttp.NewHandler(authCommand(handleGetDynamicConfigs, "config", "get"), "handleGetConfigs")).Methods("GET")
	router.Handle("/v2/configs/{bundle}/{layer}/{owner}/{key}", otelhttp.NewHandler(authCommand(handleGetDynamicConfigs, "config", "get"), "handleGetConfigs")).Methods("GET")
	router.Handle("/v2/configs/{bundle}/{layer}/{owner}/{key}", otelhttp.NewHandler(authCommand(handlePutDynamicConfiguration, "config", "set"), "handlePutDynamicConfiguration")).Methods("PUT")
	router.Handle("/v2/configs/{bundle}/{layer}/{owner}/{key}", otelhttp.NewHandler(authCommand(handleDeleteDynamicConfig, "config", "delete"), "handleDeleteConfig")).Methods("DELETE")
}
