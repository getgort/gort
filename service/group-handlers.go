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

	"github.com/getgort/gort/data/rest"
	gerrs "github.com/getgort/gort/errors"
)

// handleDeleteGroup handles "DELETE /v2/groups/{groupname}"
func handleDeleteGroup(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	err := dataAccessLayer.GroupDelete(r.Context(), params["groupname"])
	if err != nil {
		respondAndLogError(w, err)
		return
	}
}

// handleDeleteGroupMember handles "DELETE "/v2/groups/{groupname}/members/{username}""
func handleDeleteGroupMember(w http.ResponseWriter, r *http.Request) {
	var exists bool
	var err error

	params := mux.Vars(r)
	groupname := params["groupname"]
	username := params["username"]

	exists, err = dataAccessLayer.GroupExists(r.Context(), groupname)
	if err != nil {
		respondAndLogError(w, err)
		return
	}
	if !exists {
		http.Error(w, "no such group", http.StatusNotFound)
		return
	}

	exists, err = dataAccessLayer.UserExists(r.Context(), username)
	if err != nil {
		respondAndLogError(w, err)
		return
	}
	if !exists {
		http.Error(w, "no such user", http.StatusNotFound)
		return
	}

	err = dataAccessLayer.GroupRemoveUser(r.Context(), groupname, username)
	if err != nil {
		respondAndLogError(w, err)
	}
}

// handleGetGroup handles "GET /v2/groups/{groupname}"
func handleGetGroup(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	exists, err := dataAccessLayer.GroupExists(r.Context(), params["groupname"])
	if err != nil {
		respondAndLogError(w, err)
		return
	}
	if !exists {
		http.Error(w, "No such group", http.StatusNotFound)
		return
	}

	group, err := dataAccessLayer.GroupGet(r.Context(), params["groupname"])
	if err != nil {
		respondAndLogError(w, err)
		return
	}

	json.NewEncoder(w).Encode(group)
}

// handleGetGroups handles "GET /v2/groups"
func handleGetGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := dataAccessLayer.GroupList(r.Context())

	if err != nil {
		respondAndLogError(w, err)
		return
	}

	json.NewEncoder(w).Encode(groups)
}

// handleGetGroupMembers handles "GET /v2/groups/{groupname}/members"
func handleGetGroupMembers(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupname := params["groupname"]

	exists, err := dataAccessLayer.GroupExists(r.Context(), groupname)
	if err != nil {
		respondAndLogError(w, err)
		return
	}
	if !exists {
		http.Error(w, "no such group", http.StatusNotFound)
		return
	}

	group, err := dataAccessLayer.GroupGet(r.Context(), groupname)
	if err != nil {
		respondAndLogError(w, err)
		return
	}

	json.NewEncoder(w).Encode(group.Users)
}

// handlePutGroup handles "PUT /v2/groups/{groupname}"
func handlePutGroup(w http.ResponseWriter, r *http.Request) {
	var group rest.Group
	var err error

	params := mux.Vars(r)

	err = json.NewDecoder(r.Body).Decode(&group)
	if err != nil {
		respondAndLogError(w, gerrs.ErrUnmarshal)
		return
	}

	group.Name = params["groupname"]

	exists, err := dataAccessLayer.GroupExists(r.Context(), group.Name)
	if err != nil {
		respondAndLogError(w, err)
		return
	}

	// NOTE: Should we just make "update" create groups that don't exist?
	if exists {
		err = dataAccessLayer.GroupUpdate(r.Context(), group)
	} else {
		err = dataAccessLayer.GroupCreate(r.Context(), group)
	}

	if err != nil {
		respondAndLogError(w, err)
	}
}

// handlePutGroupMember handles "PUT "/v2/groups/{groupname}/members/{username}""
func handlePutGroupMember(w http.ResponseWriter, r *http.Request) {
	var exists bool
	var err error

	params := mux.Vars(r)
	groupname := params["groupname"]
	username := params["username"]

	exists, err = dataAccessLayer.GroupExists(r.Context(), groupname)
	if err != nil {
		respondAndLogError(w, err)
		return
	}
	if !exists {
		http.Error(w, "no such group", http.StatusNotFound)
		return
	}

	exists, err = dataAccessLayer.UserExists(r.Context(), username)
	if err != nil {
		respondAndLogError(w, err)
		return
	}
	if !exists {
		http.Error(w, "no such user", http.StatusNotFound)
		return
	}

	err = dataAccessLayer.GroupAddUser(r.Context(), groupname, username)
	if err != nil {
		respondAndLogError(w, err)
	}
}

func addGroupMethodsToRouter(router *mux.Router) {
	// Basic group methods
	router.HandleFunc("/v2/groups", handleGetGroups).Methods("GET")
	router.HandleFunc("/v2/groups/{groupname}", handleGetGroup).Methods("GET")
	router.HandleFunc("/v2/groups/{groupname}", handlePutGroup).Methods("PUT")
	router.HandleFunc("/v2/groups/{groupname}", handleDeleteGroup).Methods("DELETE")

	// Group user membership
	router.HandleFunc("/v2/groups/{groupname}/members", handleGetGroupMembers).Methods("GET")
	router.HandleFunc("/v2/groups/{groupname}/members/{username}", handleDeleteGroupMember).Methods("DELETE")
	router.HandleFunc("/v2/groups/{groupname}/members/{username}", handlePutGroupMember).Methods("PUT")
}
