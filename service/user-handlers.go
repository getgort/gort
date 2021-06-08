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

	"github.com/getgort/gort/data/rest"
)

// handleDeleteUser handles "DELETE /v2/users/{username}"
func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	err := dataAccessLayer.UserDelete(r.Context(), params["username"])

	if err != nil {
		respondAndLogError(w, err)
		return
	}
}

// handleDeleteUserGroup handles "DELETE /v2/users/{username}/groups/{username}"
func handleDeleteUserGroup(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not Implemented", http.StatusNotImplemented)
}

// handleGetUser handles "GET /v2/users/{username}"
func handleGetUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	exists, err := dataAccessLayer.UserExists(r.Context(), params["username"])
	if err != nil {
		respondAndLogError(w, err)
		return
	}
	if !exists {
		http.Error(w, "No such user", http.StatusNotFound)
		return
	}

	user, err := dataAccessLayer.UserGet(r.Context(), params["username"])
	if err != nil {
		respondAndLogError(w, err)
		return
	}

	json.NewEncoder(w).Encode(user)
}

// handleGetUserGroups handles "GET /v2/users/{username}/groups"
func handleGetUserGroups(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	exists, err := dataAccessLayer.UserExists(r.Context(), params["username"])
	if err != nil {
		respondAndLogError(w, err)
		return
	}
	if !exists {
		http.Error(w, "No such user", http.StatusNotFound)
		return
	}

	groups, err := dataAccessLayer.UserGroupList(r.Context(), params["username"])
	if err != nil {
		respondAndLogError(w, err)
		return
	}

	json.NewEncoder(w).Encode(groups)
}

// handleGetUsers handles "GET /v2/users"
func handleGetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := dataAccessLayer.UserList(r.Context())

	if err != nil {
		respondAndLogError(w, err)
		return
	}

	json.NewEncoder(w).Encode(users)
}

// handlePutUser handles "POST /v2/users/{username}"
func handlePutUser(w http.ResponseWriter, r *http.Request) {
	var user rest.User
	var err error

	params := mux.Vars(r)

	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	user.Username = params["username"]

	exists, err := dataAccessLayer.UserExists(r.Context(), user.Username)
	if err != nil {
		respondAndLogError(w, err)
		return
	}

	if exists {
		err = dataAccessLayer.UserUpdate(r.Context(), user)
	} else {
		err = dataAccessLayer.UserCreate(r.Context(), user)
	}

	if err != nil {
		respondAndLogError(w, err)
		return
	}
}

// handlePutUserGroup handles "PUT /v2/users/{username}/groups/{username}"
func handlePutUserGroup(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not Implemented", http.StatusNotImplemented)
}

func addUserMethodsToRouter(router *mux.Router) {
	router.Handle("/v2/users", otelhttp.NewHandler(http.HandlerFunc(handleGetUsers), "handleGetUsers")).Methods("GET")
	router.Handle("/v2/users/{username}", otelhttp.NewHandler(http.HandlerFunc(handleGetUser), "handleGetUser")).Methods("GET")
	router.Handle("/v2/users/{username}", otelhttp.NewHandler(http.HandlerFunc(handlePutUser), "handlePutUser")).Methods("PUT")
	router.Handle("/v2/users/{username}", otelhttp.NewHandler(http.HandlerFunc(handleDeleteUser), "handleDeleteUser")).Methods("DELETE")

	// User group membership
	router.Handle("/v2/users/{username}/groups", otelhttp.NewHandler(http.HandlerFunc(handleGetUserGroups), "handleGetUserGroups")).Methods("GET")
	router.Handle("/v2/users/{username}/groups/{username}", otelhttp.NewHandler(http.HandlerFunc(handleDeleteUserGroup), "handleDeleteUserGroup")).Methods("DELETE")
	router.Handle("/v2/users/{username}/groups/{username}", otelhttp.NewHandler(http.HandlerFunc(handlePutUserGroup), "handlePutUserGroup")).Methods("PUT")
}
