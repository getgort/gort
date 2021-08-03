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
	"sync"

	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// handleDeleteGroup handles "DELETE /v2/roles/{rolename}"
func handleDeleteRole(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	dataAccessLayer, err := dataaccess.Get()
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	err = dataAccessLayer.RoleDelete(r.Context(), params["rolename"])
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
}

// handleGetRoles handles "GET /v2/roles"
func handleGetRoles(w http.ResponseWriter, r *http.Request) {
	dataAccessLayer, err := dataaccess.Get()
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	roles, err := dataAccessLayer.RoleList(r.Context())

	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	json.NewEncoder(w).Encode(roles)
}

// handleGetRole handles "GET /v2/roles/{rolename}"
func handleGetRole(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	rolename := params["rolename"]

	dataAccessLayer, err := dataaccess.Get()
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	exists, err := dataAccessLayer.RoleExists(r.Context(), rolename)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
	if !exists {
		http.Error(w, "no such role", http.StatusNotFound)
		return
	}

	var errs chan error = make(chan error, 3)
	var role rest.Role
	var perms rest.RolePermissionList
	var groups []rest.Group

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		o, err := dataAccessLayer.RoleGet(r.Context(), rolename)
		if err != nil {
			errs <- err
		}

		role = o
		wg.Done()
	}()

	go func() {
		o, err := dataAccessLayer.RolePermissionList(r.Context(), rolename)
		if err != nil {
			errs <- err
		}

		perms = o
		wg.Done()
	}()

	go func() {
		o, err := dataAccessLayer.RoleGroupList(r.Context(), rolename)
		if err != nil {
			errs <- err
		}

		groups = o
		wg.Done()
	}()

	wg.Wait()

	close(errs)

	if err := <-errs; err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	role.Permissions = perms
	role.Groups = groups

	json.NewEncoder(w).Encode(role)
}

// handleGetRole handles "GET /v2/roles/{rolename}"
func handleGetRolePermissions(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	rolename := params["rolename"]

	dataAccessLayer, err := dataaccess.Get()
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	exists, err := dataAccessLayer.RoleExists(r.Context(), rolename)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
	if !exists {
		http.Error(w, "no such role", http.StatusNotFound)
		return
	}

	rpl, err := dataAccessLayer.RolePermissionList(r.Context(), rolename)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	json.NewEncoder(w).Encode(rpl)
}

// handleGrantRolePermission handles "PUT /v2/roles/{rolename}/bundles/{bundlename}/permissions/{permissionname}"
func handleGrantRolePermission(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	rolename := params["rolename"]
	bundlename := params["bundlename"]
	permissionname := params["permissionname"]

	dataAccessLayer, err := dataaccess.Get()
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	exists, err := dataAccessLayer.RoleExists(r.Context(), rolename)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
	if !exists {
		http.Error(w, "no such role", http.StatusNotFound)
		return
	}

	err = dataAccessLayer.RolePermissionAdd(r.Context(), rolename, bundlename, permissionname)
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

	dataAccessLayer, err := dataaccess.Get()
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	exists, err := dataAccessLayer.RoleExists(r.Context(), rolename)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
	if !exists {
		http.Error(w, "no such role", http.StatusNotFound)
		return
	}

	err = dataAccessLayer.RolePermissionDelete(r.Context(), rolename, bundlename, permissionname)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
}

// handlePutRole handles "PUT /v2/roles/{rolename}"
func handlePutRole(w http.ResponseWriter, r *http.Request) {
	var err error

	params := mux.Vars(r)

	dataAccessLayer, err := dataaccess.Get()
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	err = dataAccessLayer.RoleCreate(r.Context(), params["rolename"])
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
}

func addRoleMethodsToRouter(router *mux.Router) {
	// Basic role methods
	router.Handle("/v2/roles", otelhttp.NewHandler(authCommand(handleGetRoles, "role", "list"), "handleGetRoles")).Methods("GET")
	router.Handle("/v2/roles/{rolename}", otelhttp.NewHandler(authCommand(handleGetRole, "role", "info"), "handleGetRole")).Methods("GET")
	router.Handle("/v2/roles/{rolename}", otelhttp.NewHandler(authCommand(handlePutRole, "role", "create"), "handlePutRole")).Methods("PUT")
	router.Handle("/v2/roles/{rolename}", otelhttp.NewHandler(authCommand(handleDeleteRole, "role", "delete"), "handleDeleteRole")).Methods("DELETE")

	// Role permissions
	router.Handle("/v2/roles/{rolename}/permissions", otelhttp.NewHandler(authCommand(handleGetRolePermissions, "role", "info"), "handleGetRolePermissions")).Methods("GET")
	router.Handle("/v2/roles/{rolename}/bundles/{bundlename}/permissions/{permissionname}", otelhttp.NewHandler(authCommand(handleRevokeRolePermission, "role", "revoke-permission"), "handleDeleteRolePermission")).Methods("DELETE")
	router.Handle("/v2/roles/{rolename}/bundles/{bundlename}/permissions/{permissionname}", otelhttp.NewHandler(authCommand(handleGrantRolePermission, "role", "grant-permission"), "handlePutRolePermission")).Methods("PUT")
}
