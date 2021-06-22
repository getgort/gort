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

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// handleDeleteGroup handles "DELETE /v2/roles/{rolename}"
func handleDeleteRole(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	err := dataAccessLayer.RoleDelete(r.Context(), params["rolename"])
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
}

// handleGetRole handles "GET /v2/roles/{rolename}"
func handleGetRole(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	rolename := params["rolename"]

	exists, err := dataAccessLayer.RoleExists(r.Context(), rolename)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
	if !exists {
		http.Error(w, "no such role", http.StatusNotFound)
		return
	}

	role, err := dataAccessLayer.RoleGet(r.Context(), rolename)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	json.NewEncoder(w).Encode(role)
}

// handleGrantRolePermission handles "PUT /v2/roles/{rolename}/bundles/{bundlename}/permissions/{permissionname}"
func handleGrantRolePermission(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	rolename := params["rolename"]
	bundlename := params["bundlename"]
	permissionname := params["permissionname"]

	exists, err := dataAccessLayer.RoleExists(r.Context(), rolename)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
	if !exists {
		http.Error(w, "no such role", http.StatusNotFound)
		return
	}

	err = dataAccessLayer.RoleGrantPermission(r.Context(), rolename, bundlename, permissionname)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
}

// handleRevokeRolePermission handles "DELETE /v2/roles/{rolename}/bundles/{bundlename}/permissions/{permissionname}"
func handleRevokeRolePermission(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	rolename := params["rolename"]
	bundlename := params["bundlename"]
	permissionname := params["permissionname"]

	exists, err := dataAccessLayer.RoleExists(r.Context(), rolename)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
	if !exists {
		http.Error(w, "no such role", http.StatusNotFound)
		return
	}

	err = dataAccessLayer.RoleRevokePermission(r.Context(), rolename, bundlename, permissionname)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
}

// handlePutRole handles "PUT /v2/roles/{rolename}"
func handlePutRole(w http.ResponseWriter, r *http.Request) {
	var err error

	params := mux.Vars(r)

	err = dataAccessLayer.RoleCreate(r.Context(), params["rolename"])
	if err != nil {
		respondAndLogError(r.Context(), w, err)
	}
}

func addRoleMethodsToRouter(router *mux.Router) {
	// Basic role methods
	router.Handle("/v2/roles/{rolename}", otelhttp.NewHandler(http.HandlerFunc(handleGetRole), "handleGetRole")).Methods("GET")
	router.Handle("/v2/roles/{rolename}", otelhttp.NewHandler(http.HandlerFunc(handlePutRole), "handlePutRole")).Methods("PUT")
	router.Handle("/v2/roles/{rolename}", otelhttp.NewHandler(http.HandlerFunc(handleDeleteRole), "handleDeleteRole")).Methods("DELETE")

	// Role permissions
	router.Handle("/v2/roles/{rolename}/bundles/{bundlename}/permissions/{permissionname}", otelhttp.NewHandler(http.HandlerFunc(handleRevokeRolePermission), "handleDeleteRolePermission")).Methods("DELETE")
	router.Handle("/v2/roles/{rolename}/bundles/{bundlename}/permissions/{permissionname}", otelhttp.NewHandler(http.HandlerFunc(handleGrantRolePermission), "handlePutRolePermission")).Methods("PUT")
}
