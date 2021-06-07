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
	"errors"
	"net/http"
	"sort"
	"strings"

	"github.com/gorilla/mux"

	"github.com/getgort/gort/config"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess/errs"
	gerrs "github.com/getgort/gort/errors"
)

var (
	// ErrMissingValue is returned by a method when an expected form field
	// is missing.
	ErrMissingValue = errors.New("a form value is missing")
)

// handleGetBundles handles "GET /v2/bundles"
func handleGetBundles(w http.ResponseWriter, r *http.Request) {
	bundles, err := getAllBundles()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if len(bundles) == 0 {
		http.Error(w, "No bundles found", http.StatusNoContent)
		return
	}

	json.NewEncoder(w).Encode(bundles)
}

// handleGetBundleVersions handles "GET /v2/bundles/{name}/versions"
func handleGetBundleVersions(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]

	bundles, err := dataAccessLayer.BundleListVersions(name)
	if err != nil {
		respondAndLogError(w, err)
		return
	} else if len(bundles) == 0 {
		http.Error(w, "No bundles found", http.StatusNoContent)
		return
	}

	json.NewEncoder(w).Encode(bundles)
}

// handleDeleteBundleVersion handles "DELETE /v2/bundles/{name}/versions/{version}"
func handleDeleteBundleVersion(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]
	version := params["version"]

	err := dataAccessLayer.BundleDelete(name, version)
	if err != nil {
		respondAndLogError(w, err)
		return
	}
}

// handleGetBundleVersion handles "GET /v2/bundles/{name}/versions/{version}"
func handleGetBundleVersion(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]
	version := params["version"]

	bundle, err := dataAccessLayer.BundleGet(name, version)
	if err != nil {
		respondAndLogError(w, err)
		return
	}

	json.NewEncoder(w).Encode(bundle)
}

// handlePatchBundleVersion handles "PATCH /v2/bundles/{name}/versions/{version}"
func handlePatchBundleVersion(w http.ResponseWriter, r *http.Request) {
	var err error

	params := mux.Vars(r)
	name := params["name"]
	version := params["version"]

	enabledValue := strings.ToUpper(r.FormValue("enabled"))
	if enabledValue == "" {
		enabledValue = "-"
	}

	// If enabled=false we ignore the value of version and replace it with the
	// current enabled version.
	if enabledValue[0] == 'F' {
		version, err = dataAccessLayer.BundleEnabledVersion(name)
		if err != nil {
			respondAndLogError(w, err)
			return
		}
	}

	exists, err := dataAccessLayer.BundleExists(name, version)
	if err != nil {
		respondAndLogError(w, err)
		return
	}
	if !exists {
		respondAndLogError(w, errs.ErrNoSuchBundle)
		return
	}

	if r.ContentLength > 0 {
		var bundle data.Bundle

		err = json.NewDecoder(r.Body).Decode(&bundle)
		if err != nil {
			respondAndLogError(w, gerrs.ErrUnmarshal)
			return
		}

		bundle.Name = name
		bundle.Version = version

		err = dataAccessLayer.BundleUpdate(bundle)
		if err != nil {
			respondAndLogError(w, err)
			return
		}
	}

	if enabledValue[0] == 'T' {
		err = dataAccessLayer.BundleEnable(name, version)
	} else if enabledValue[0] == 'F' {
		err = dataAccessLayer.BundleDisable(name)
	}
	if err != nil {
		respondAndLogError(w, err)
		return
	}
}

// handlePutBundleVersion handles "PUT /v2/bundles/{name}/versions/{version}"
func handlePutBundleVersion(w http.ResponseWriter, r *http.Request) {
	var bundle data.Bundle
	var err error

	err = json.NewDecoder(r.Body).Decode(&bundle)
	if err != nil {
		respondAndLogError(w, gerrs.ErrUnmarshal)
		return
	}

	params := mux.Vars(r)
	bundle.Name = params["name"]
	bundle.Version = params["version"]

	err = dataAccessLayer.BundleCreate(bundle)
	if err != nil {
		respondAndLogError(w, err)
		return
	}
}

func getAllBundles() ([]data.Bundle, error) {
	// Default bundles from config
	bundles := config.GetBundleConfigs()

	// Explicit bundles from the data layer
	dataLayerBundles, err := dataAccessLayer.BundleList()
	if err != nil {
		return nil, err
	}

	bundles = append(bundles, dataLayerBundles...)

	sort.Slice(bundles, func(i, j int) bool { return bundles[i].Name < bundles[j].Name })

	return bundles, nil
}

func addBundleMethodsToRouter(router *mux.Router) {
	router.HandleFunc("/v2/bundles", handleGetBundles).Methods("GET")

	router.HandleFunc("/v2/bundles/{name}", handleGetBundleVersions).Methods("GET")
	router.HandleFunc("/v2/bundles/{name}/versions", handleGetBundleVersions).Methods("GET")

	router.HandleFunc("/v2/bundles/{name}/versions/{version}", handleGetBundleVersion).Methods("GET")
	router.HandleFunc("/v2/bundles/{name}/versions/{version}", handlePutBundleVersion).Methods("PUT")
	router.HandleFunc("/v2/bundles/{name}/versions/{version}", handleDeleteBundleVersion).Methods("DELETE")

	router.HandleFunc("/v2/bundles/{name}/versions/{version}", handlePatchBundleVersion).Methods("PATCH")
	router.HandleFunc("/v2/bundles/{name}/versions/{version}", handlePatchBundleVersion).Methods("PATCH").
		Queries("enabled", "")
}
