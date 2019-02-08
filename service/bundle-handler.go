package service

import (
	"encoding/json"
	"net/http"

	"github.com/clockworksoul/cog2/data"
	cogerr "github.com/clockworksoul/cog2/errors"
	"github.com/gorilla/mux"
)

// handleGetBundles handles "GET /v2/bundles"
func handleGetBundles(w http.ResponseWriter, r *http.Request) {
	bundles, err := dataAccessLayer.BundleList()
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

// handlePutBundleVersion handles "PUT /v2/bundles/{name}/versions/{version}"
func handlePutBundleVersion(w http.ResponseWriter, r *http.Request) {
	var bundle data.Bundle
	var err error

	err = json.NewDecoder(r.Body).Decode(&bundle)
	if err != nil {
		respondAndLogError(w, cogerr.ErrUnmarshal)
		return
	}

	params := mux.Vars(r)
	bundle.Name = params["name"]
	bundle.Version = params["version"]

	err = dataAccessLayer.BundleCreate(bundle)
	if err != nil {
		respondAndLogError(w, err)
	}
}

func addBundleMethodsToRouter(router *mux.Router) {
	router.HandleFunc("/v2/bundles", handleGetBundles).Methods("GET")

	router.HandleFunc("/v2/bundles/{name}", handleGetBundleVersions).Methods("GET")
	router.HandleFunc("/v2/bundles/{name}/versions", handleGetBundleVersions).Methods("GET")

	router.HandleFunc("/v2/bundles/{name}/versions/{version}", handleGetBundleVersion).Methods("GET")
	router.HandleFunc("/v2/bundles/{name}/versions/{version}", handlePutBundleVersion).Methods("PUT")
	router.HandleFunc("/v2/bundles/{name}/versions/{version}", handleDeleteBundleVersion).Methods("DELETE")
}
