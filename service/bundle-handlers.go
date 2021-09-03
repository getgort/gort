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
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess"
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
	bundles, err := getAllBundles(r.Context())

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

	dataAccessLayer, err := dataaccess.Get()
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	bundles, err := dataAccessLayer.BundleVersionList(r.Context(), name)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
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

	dataAccessLayer, err := dataaccess.Get()
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	err = dataAccessLayer.BundleDelete(r.Context(), name, version)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
}

// handleGetBundleVersion handles "GET /v2/bundles/{name}/versions/{version}"
func handleGetBundleVersion(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]
	version := params["version"]

	dataAccessLayer, err := dataaccess.Get()
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	bundle, err := dataAccessLayer.BundleGet(r.Context(), name, version)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
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

	dataAccessLayer, err := dataaccess.Get()
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	// If enabled=false we ignore the value of version and replace it with the
	// current enabled version.
	if enabledValue[0] == 'F' {
		version, err = dataAccessLayer.BundleEnabledVersion(r.Context(), name)
		if err != nil {
			respondAndLogError(r.Context(), w, err)
			return
		}
	}

	exists, err := dataAccessLayer.BundleExists(r.Context(), name, version)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
	if !exists {
		respondAndLogError(r.Context(), w, errs.ErrNoSuchBundle)
		return
	}

	if r.ContentLength > 0 {
		var bundle data.Bundle

		err = json.NewDecoder(r.Body).Decode(&bundle)
		if err != nil {
			respondAndLogError(r.Context(), w, gerrs.ErrUnmarshal)
			return
		}

		bundle.Name = name
		bundle.Version = version

		err = dataAccessLayer.BundleUpdate(r.Context(), bundle)
		if err != nil {
			respondAndLogError(r.Context(), w, err)
			return
		}
	}

	if enabledValue[0] == 'T' {
		err = dataAccessLayer.BundleEnable(r.Context(), name, version)
	} else if enabledValue[0] == 'F' {
		err = dataAccessLayer.BundleDisable(r.Context(), name, version)
	}
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
}

// handlePutBundleVersion handles "PUT /v2/bundles/{name}/versions/{version}"
func handlePutBundleVersion(w http.ResponseWriter, r *http.Request) {
	var bundle data.Bundle
	var err error

	err = json.NewDecoder(r.Body).Decode(&bundle)
	if err != nil {
		respondAndLogError(r.Context(), w, gerrs.ErrUnmarshal)
		return
	}

	params := mux.Vars(r)
	bundle.Name = params["name"]
	bundle.Version = params["version"]

	dataAccessLayer, err := dataaccess.Get()
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	err = dataAccessLayer.BundleCreate(r.Context(), bundle)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
}

func getAllBundles(ctx context.Context) ([]data.Bundle, error) {
	dataAccessLayer, err := dataaccess.Get()
	if err != nil {
		return nil, err
	}

	// Explicit bundles from the data layer
	bundles, err := dataAccessLayer.BundleList(ctx)
	if err != nil {
		return nil, err
	}

	sort.Slice(bundles, func(i, j int) bool { return bundles[i].Name < bundles[j].Name })

	return bundles, nil
}

func addBundleMethodsToRouter(router *mux.Router) {
	router.Handle("/v2/bundles", otelhttp.NewHandler(authCommand(handleGetBundles, "bundle", "info"), "handleGetBundles")).Methods("GET")

	router.Handle("/v2/bundles/{name}", otelhttp.NewHandler(authCommand(handleGetBundleVersions, "bundle", "info"), "handleGetBundleVersions")).Methods("GET")
	router.Handle("/v2/bundles/{name}/versions", otelhttp.NewHandler(authCommand(handleGetBundleVersions, "bundle", "list"), "handleGetBundleVersions")).Methods("GET")

	router.Handle("/v2/bundles/{name}/versions/{version}", otelhttp.NewHandler(authCommand(handleGetBundleVersion, "bundle", "info"), "handleGetBundleVersion")).Methods("GET")
	router.Handle("/v2/bundles/{name}/versions/{version}", otelhttp.NewHandler(authCommand(handlePutBundleVersion, "bundle", "install"), "handlePutBundleVersion")).Methods("PUT")
	router.Handle("/v2/bundles/{name}/versions/{version}", otelhttp.NewHandler(authCommand(handleDeleteBundleVersion, "bundle", "install"), "handleDeleteBundleVersion")).Methods("DELETE")

	router.Handle("/v2/bundles/{name}/versions/{version}", otelhttp.NewHandler(authCommand(handlePatchBundleVersion, "bundle", "enable"), "handlePatchBundleVersion")).Methods("PATCH")
	router.Handle("/v2/bundles/{name}/versions/{version}", otelhttp.NewHandler(authCommand(handlePatchBundleVersion, "bundle", "enable"), "handlePatchBundleVersion")).Methods("PATCH").Queries("enabled", "")
}
