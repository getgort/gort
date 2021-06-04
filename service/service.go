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
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess"
	"github.com/getgort/gort/dataaccess/errs"
	gerrs "github.com/getgort/gort/errors"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var (
	dataAccessLayer dataaccess.DataAccess
)

// RequestEvent represents a request of a service endpoint.
type RequestEvent struct {
	Addr      string
	UserID    string
	Timestamp time.Time
	Request   string
	Status    int
	Size      int64
}

func (e RequestEvent) String() string {
	const dateFormat = "02/Jan/2006:15:04:05 -0700"

	return fmt.Sprintf("%s - %s [%v] %q %d %d",
		e.Addr,
		e.UserID,
		e.Timestamp.Format(dateFormat),
		e.Request,
		e.Status,
		e.Size,
	)
}

// StatusCaptureWriter is a wrapper around a http.ResponseWriter that is used
// by middleware to capture a response status and byte length for logging
// purposes.
type StatusCaptureWriter struct {
	http.ResponseWriter

	status *int
	bytes  *int
}

// Header returns the header map that will be sent by http.WriteHeader.
func (w StatusCaptureWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// Write writes the data to the connection as part of an HTTP reply.
func (w StatusCaptureWriter) Write(bytes []byte) (int, error) {
	*w.bytes = len(bytes)
	return w.ResponseWriter.Write(bytes)
}

// WriteHeader sends an HTTP response header with the provided status code.
func (w StatusCaptureWriter) WriteHeader(statusCode int) {
	*w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func bootstrapUserWithDefaults(user rest.User) (rest.User, error) {
	// If user doesn't have a defined email, we default to "gort@localhost".
	if user.Email == "" {
		user.Email = "gort@localhost"
	}

	// If user doesn't have a defined name, we default to "Gort Administrator".
	if user.FullName == "" {
		user.FullName = "Gort Administrator"
	}

	// The bootstrap user is _always_ named "admin".
	user.Username = "admin"

	// If user doesn't have a defined password, we kindly generate one.
	if user.Password == "" {
		password, err := data.GenerateRandomToken(32)
		if err != nil {
			return user, err
		}
		user.Password = password
	}

	return user, nil
}

func buildLoggingMiddleware(logsous chan RequestEvent) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			status := 200
			bytelen := 0

			// Call the next handler, which can be another middleware in the chain, or the final handler.
			next.ServeHTTP(StatusCaptureWriter{w, &status, &bytelen}, r)

			// If there's a token, retrieve it for logging purposes.
			userID := "-"
			tokenString := r.Header.Get("X-Session-Token")
			if tokenString != "" {
				token, _ := dataAccessLayer.TokenRetrieveByToken(tokenString)
				userID = token.User
			}

			requestLine := fmt.Sprintf("%s %s %s",
				r.Method,
				r.RequestURI,
				r.Proto)

			e := RequestEvent{
				Addr:      r.RemoteAddr,
				UserID:    userID,
				Timestamp: time.Now(),
				Request:   requestLine,
				Status:    status,
				Size:      int64(bytelen),
			}

			logsous <- e
		})
	}
}

func tokenObservingMiddleware(next http.Handler) http.Handler {
	exemptEndpoints := map[string]bool{
		"/v2/authenticate": true,
		"/v2/bootstrap":    true,
		"/v2/healthz":      true,
		"/v2/metrics":      true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestURI := strings.Split(r.RequestURI, "?")[0]

		if exemptEndpoints[requestURI] {
			next.ServeHTTP(w, r)
			return
		}

		token := r.Header.Get("X-Session-Token")
		if token == "" || !dataAccessLayer.TokenEvaluate(token) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RESTServer represents a Gort REST API service.
type RESTServer struct {
	*http.Server

	requests chan RequestEvent
}

// BuildRESTServer builds a RESTServer.
func BuildRESTServer(addr string) *RESTServer {
	dalUpdate := dataaccess.Updates()

	for dalState := range dalUpdate {
		if dalState == dataaccess.StateInitialized {
			break
		}
	}

	var err error
	dataAccessLayer, err = dataaccess.Get()
	if err != nil {
		log.WithError(err).Fatal("Could not connect to data access layer")
	}

	requests := make(chan RequestEvent)

	router := mux.NewRouter()
	router.Use(buildLoggingMiddleware(requests), tokenObservingMiddleware)

	addHealthzMethodToRouter(router)
	addBundleMethodsToRouter(router)
	addGroupMethodsToRouter(router)
	addUserMethodsToRouter(router)

	server := &http.Server{Addr: addr, Handler: router}

	return &RESTServer{server, requests}
}

// Requests retrieves the channel to which user request events are sent.
func (s *RESTServer) Requests() <-chan RequestEvent {
	return s.requests
}

// ListenAndServe starts the Gort web service.
func (s *RESTServer) ListenAndServe() error {
	log.WithField("address", s.Addr).Info("Gort controller is starting")

	return s.Server.ListenAndServe()
}

// handleAuthenticate handles "GET /authenticate"
func handleAuthenticate(w http.ResponseWriter, r *http.Request) {
	// Grab the user struct from the request. If it doesn't exist, respond with
	// a client error.
	user := rest.User{}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		respondAndLogError(w, gerrs.ErrUnmarshal)
		return
	}

	username := user.Username
	password := user.Password

	le := log.WithField("user", user)

	exists, err := dataAccessLayer.UserExists(username)
	if err != nil {
		le.WithError(err).Error("Authentication: failed to find user")
		return
	}

	if !exists {
		http.Error(w, "No such user", http.StatusBadRequest)
		le.Error("Authentication: No such user")
		return
	}

	authenticated, err := dataAccessLayer.UserAuthenticate(username, password)
	if err != nil {
		respondAndLogError(w, err)
		return
	}

	if !authenticated {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	token, err := dataAccessLayer.TokenGenerate(username, 10*time.Minute)
	if err != nil {
		respondAndLogError(w, err)
		return
	}

	json.NewEncoder(w).Encode(token)
}

func respondAndLogError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	msg := err.Error()

	switch {
	// A required field is empty or missing
	case gerrs.Is(err, errs.ErrEmptyBundleName):
		fallthrough
	case gerrs.Is(err, errs.ErrEmptyBundleVersion):
		fallthrough
	case gerrs.Is(err, errs.ErrEmptyGroupName):
		fallthrough
	case gerrs.Is(err, errs.ErrEmptyUserName):
		fallthrough
	case gerrs.Is(err, ErrMissingValue):
		fallthrough
	case gerrs.Is(err, errs.ErrFieldRequired):
		status = http.StatusExpectationFailed

	// Requested resource doesn't exist
	case gerrs.Is(err, errs.ErrNoSuchBundle):
		fallthrough
	case gerrs.Is(err, errs.ErrNoSuchGroup):
		fallthrough
	case gerrs.Is(err, errs.ErrNoSuchToken):
		fallthrough
	case gerrs.Is(err, errs.ErrNoSuchUser):
		status = http.StatusNotFound

	// Nope
	case gerrs.Is(err, errs.ErrAdminUndeletable):
		status = http.StatusForbidden

	// Can't insert over something that already exists
	case gerrs.Is(err, errs.ErrBundleExists):
		fallthrough
	case gerrs.Is(err, errs.ErrGroupExists):
		fallthrough
	case gerrs.Is(err, errs.ErrUserExists):
		status = http.StatusConflict

	// Not done yet
	case gerrs.Is(err, errs.ErrNotImplemented):
		status = http.StatusNotImplemented

	// Data access errors
	case gerrs.Is(err, errs.ErrDataAccessNotInitialized):
		fallthrough
	case gerrs.Is(err, errs.ErrDataAccessCantInitialize):
		fallthrough
	case gerrs.Is(err, errs.ErrDataAccessCantConnect):
		fallthrough
	case gerrs.Is(err, errs.ErrDataAccess):
		status = http.StatusInternalServerError
		log.WithField("status", status).Error(msg)

	// Bad context
	case gerrs.Is(err, gerrs.ErrUnmarshal):
		msg = "Corrupt JSON payload"
		status = http.StatusNotAcceptable

	// Something else?
	default:
		log.WithError(err).Warn("Unhandled server error")
		status = http.StatusInternalServerError
		log.WithError(err).WithField("status", status).Error(msg)
	}

	http.Error(w, msg, status)
}

// handleBootstrap handles "POST /bootstrap"
func handleBootstrap(w http.ResponseWriter, r *http.Request) {
	users, err := dataAccessLayer.UserList()
	if err != nil {
		respondAndLogError(w, err)
		return
	}

	// If we already have users on this host, reject as "already bootstrapped".
	if len(users) != 0 {
		http.Error(w, "Service already bootstrapped", http.StatusConflict)
		log.Warn("Re-bootstrap attempted")
		return
	}

	// Grab the user struct from the request. If it doesn't exist, respond with
	// a client error.
	user := rest.User{}
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		respondAndLogError(w, gerrs.ErrUnmarshal)
		return
	}

	// Set user defaults where necessary.
	user, err = bootstrapUserWithDefaults(user)
	if err != nil {
		respondAndLogError(w, err)
		return
	}

	// Persist our shiny new user to the database.
	err = dataAccessLayer.UserCreate(user)
	if err != nil {
		respondAndLogError(w, err)
		return
	}

	// Create admin group.
	group := rest.Group{Name: "admin"}
	err = dataAccessLayer.GroupCreate(group)
	if err != nil {
		respondAndLogError(w, err)
		return
	}

	// Add the admin user to the admin group.
	err = dataAccessLayer.GroupAddUser(group.Name, user.Username)
	if err != nil {
		respondAndLogError(w, err)
		return
	}

	json.NewEncoder(w).Encode(user)
}

// handleHealthz handles "GET /healthz"
// TODO Can we make this more meaningful?
func handleHealthz(w http.ResponseWriter, r *http.Request) {
	m := map[string]bool{"healthy": true}

	json.NewEncoder(w).Encode(m)
}

func addHealthzMethodToRouter(router *mux.Router) {
	router.HandleFunc("/v2/authenticate", handleAuthenticate).Methods("POST")
	router.HandleFunc("/v2/bootstrap", handleBootstrap).Methods("POST")
	router.HandleFunc("/v2/healthz", handleHealthz).Methods("GET")
}
