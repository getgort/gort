package service

import (
	"encoding/json"
	"net/http"

	"github.com/clockworksoul/cog2/data/rest"
	"github.com/gorilla/mux"
)

// handlePostUser handles "DELETE /v2/user/{username}"
func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	err := dataAccessLayer.UserDelete(params["username"])

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handlePostUser handles "DELETE /v2/user/{username}/group/{username}"
func handleDeleteUserGroup(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not Implemented", http.StatusNotImplemented)
	return
}

// handleGetUser handles "GET /v2/user/{username}"
func handleGetUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	exists, err := dataAccessLayer.UserExists(params["username"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "No such user", http.StatusNotFound)
		return
	}

	user, err := dataAccessLayer.UserGet(params["username"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}

// handlePostUser handles "GET /v2/user/{username}/group"
func handleGetUserGroups(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	exists, err := dataAccessLayer.UserExists(params["username"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "No such user", http.StatusNotFound)
		return
	}

	groups, err := dataAccessLayer.UserGroupList(params["username"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(groups)
}

// handleGetUsers handles "GET /v2/user"
func handleGetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := dataAccessLayer.UserList()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}

// handlePostUser handles "POST /v2/user/{username}"
func handlePostUser(w http.ResponseWriter, r *http.Request) {
	var user rest.User
	var err error

	params := mux.Vars(r)

	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	user.Username = params["username"]

	exists, err := dataAccessLayer.UserExists(user.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if exists {
		err = dataAccessLayer.UserUpdate(user)
	} else {
		err = dataAccessLayer.UserCreate(user)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handlePostUser handles "PUT /v2/user/{username}/group/{username}"
func handlePutUserGroup(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not Implemented", http.StatusNotImplemented)
	return
}

func addUserMethodsToRouter(router *mux.Router) {
	router.HandleFunc("/v2/user", handleGetUsers).Methods("GET")
	router.HandleFunc("/v2/user/{username}", handleGetUser).Methods("GET")
	router.HandleFunc("/v2/user/{username}", handlePostUser).Methods("POST")
	router.HandleFunc("/v2/user/{username}", handleDeleteUser).Methods("DELETE")

	// User group membership
	router.HandleFunc("/v2/user/{username}/group", handleGetUserGroups).Methods("GET")
	router.HandleFunc("/v2/user/{username}/group/{username}", handleDeleteUserGroup).Methods("DELETE")
	router.HandleFunc("/v2/user/{username}/group/{username}", handlePutUserGroup).Methods("PUT")
}
