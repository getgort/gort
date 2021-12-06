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
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/getgort/gort/auth"
	"github.com/getgort/gort/bundles"
	"github.com/getgort/gort/config"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess"
	"github.com/getgort/gort/dataaccess/errs"
	gerrs "github.com/getgort/gort/errors"
	"github.com/getgort/gort/rules"
	"github.com/getgort/gort/telemetry"
	"github.com/getgort/gort/types"
)

var (
	ErrUnauthorized = errors.New("unauthorized")

	ErrNoSuchCommand = errors.New("no such command")

	ErrGortBundleDisabled = errors.New("gort bundle disabled")
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

// RESTServer represents a Gort REST API service.
type RESTServer struct {
	*http.Server

	requests chan RequestEvent
}

// BuildRESTServer builds a RESTServer.
func BuildRESTServer(ctx context.Context, addr string) *RESTServer {
	_, err := dataaccess.Get()
	if err != nil {
		log.WithError(err).Fatal("Could not connect to data access layer")
		telemetry.Errors().WithError(err).Commit(ctx)
	}

	requests := make(chan RequestEvent)

	router := mux.NewRouter()
	router.Use(buildLoggingMiddleware(requests), tokenObservingMiddleware)

	err = addMetricsToRouter(router)
	if err != nil {
		log.WithError(err).Fatal("Failed to add metrics endpoint to controller router")
		telemetry.Errors().WithError(err).Commit(ctx)
	}

	addAllMethodsToRouter(router)

	server := &http.Server{Addr: addr, Handler: router}

	return &RESTServer{server, requests}
}

func addAllMethodsToRouter(router *mux.Router) {
	addHealthzMethodToRouter(router)
	addBundleMethodsToRouter(router)
	addGroupMethodsToRouter(router)
	addRoleMethodsToRouter(router)
	addUserMethodsToRouter(router)
	addManagementMethodsToRouter(router)
}

// Requests retrieves the channel to which user request events are sent.
func (s *RESTServer) Requests() <-chan RequestEvent {
	return s.requests
}

// ListenAndServe starts the Gort web service.
func (s *RESTServer) ListenAndServe() error {
	log.WithField("address", s.Addr).Info("Gort controller is starting in HTTP mode")

	return s.Server.ListenAndServe()
}

// ListenAndServe starts the Gort web service.
func (s *RESTServer) ListenAndServeTLS(certFile string, keyFile string) error {
	log.WithField("address", s.Addr).Info("Gort controller is starting in HTTPS mode")

	return s.Server.ListenAndServeTLS(certFile, keyFile)
}

func addManagementMethodsToRouter(router *mux.Router) {
	router.Handle("/v2/reload", otelhttp.NewHandler(http.HandlerFunc(handleReload), "reload")).Methods("GET")
}

func addHealthzMethodToRouter(router *mux.Router) {
	router.Handle("/v2/authenticate", otelhttp.NewHandler(http.HandlerFunc(handleAuthenticate), "authenticate")).Methods("POST")
	router.Handle("/v2/bootstrap", otelhttp.NewHandler(http.HandlerFunc(handleBootstrap), "bootstrap")).Methods("POST")
	router.Handle("/v2/healthz", otelhttp.NewHandler(http.HandlerFunc(handleHealthz), "healthz")).Methods("GET")
}

func addMetricsToRouter(router *mux.Router) error {
	router.Handle("/v2/metrics", telemetry.PrometheusExporter)
	return nil
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
				dataAccessLayer, err := dataaccess.Get()
				if err != nil {
					log.WithError(err).Error(errs.ErrDataAccess)
					telemetry.Errors().WithError(err).Commit(r.Context())
					respondAndLogError(r.Context(), w, errs.ErrDataAccess)
					return
				}

				token, _ := dataAccessLayer.TokenRetrieveByToken(r.Context(), tokenString)
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

// handleAuthenticate handles "GET /authenticate"
func handleAuthenticate(w http.ResponseWriter, r *http.Request) {
	// Grab the user struct from the request. If it doesn't exist, respond with
	// a client error.
	user := rest.User{}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		respondAndLogError(r.Context(), w, gerrs.ErrUnmarshal)
		return
	}

	username := user.Username
	password := user.Password

	le := log.WithField("user", user)

	dataAccessLayer, err := dataaccess.Get()
	if err != nil {
		le.WithError(err).Error(err.Error())
		telemetry.Errors().WithError(err).Commit(r.Context())
		return
	}

	exists, err := dataAccessLayer.UserExists(r.Context(), username)
	if err != nil {
		le.WithError(err).Error("Authentication: failed to find user")
		telemetry.Errors().WithError(err).Commit(r.Context())
		return
	}

	if !exists {
		http.Error(w, "No such user", http.StatusBadRequest)
		le.Error("Authentication: No such user")
		telemetry.Errors().WithError(fmt.Errorf("no such user")).Commit(r.Context())
		return
	}

	authenticated, err := dataAccessLayer.UserAuthenticate(r.Context(), username, password)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	if !authenticated {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	token, err := dataAccessLayer.TokenGenerate(r.Context(), username, 10*time.Second)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	json.NewEncoder(w).Encode(token)
}

func DoBootstrap(ctx context.Context, user rest.User) (rest.User, error) {
	const adminGroup = "admin"
	const adminRole = "admin"
	var adminPermissions = []string{
		"manage_commands",
		"manage_groups",
		"manage_roles",
		"manage_users",
	}

	dataAccessLayer, err := dataaccess.Get()
	if err != nil {
		return user, err
	}

	// Set user defaults where necessary.
	user, err = bootstrapUserWithDefaults(user)
	if err != nil {
		return user, err
	}

	// Persist our shiny new user to the database.
	err = dataAccessLayer.UserCreate(ctx, user)
	if err != nil {
		return user, err
	}

	// Create admin group.
	err = dataAccessLayer.GroupCreate(ctx, rest.Group{Name: adminGroup})
	if err != nil {
		return user, err
	}

	// Add the admin user to the admin group.
	err = dataAccessLayer.GroupUserAdd(ctx, adminGroup, user.Username)
	if err != nil {
		return user, err
	}

	// Create an admin role
	err = dataAccessLayer.RoleCreate(ctx, adminRole)
	if err != nil {
		return user, err
	}

	// Add role to group
	err = dataAccessLayer.GroupRoleAdd(ctx, adminGroup, adminRole)
	if err != nil {
		return user, err
	}

	// Add the default permissions.
	for _, p := range adminPermissions {
		err = dataAccessLayer.RolePermissionAdd(ctx, adminRole, "gort", p)
		if err != nil {
			return user, err
		}
	}

	// Finally, add and enable the default bundle
	b, err := bundles.Default()
	if err != nil {
		return user, err
	}

	err = dataAccessLayer.BundleCreate(ctx, b)
	if err != nil {
		return user, err
	}

	err = dataAccessLayer.BundleEnable(ctx, b.Name, b.Version)
	if err != nil {
		return user, err
	}

	return user, nil
}

// handleBootstrap handles "POST /bootstrap"
func handleBootstrap(w http.ResponseWriter, r *http.Request) {
	dataAccessLayer, err := dataaccess.Get()
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	users, err := dataAccessLayer.UserList(r.Context())
	if err != nil {
		respondAndLogError(r.Context(), w, err)
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
		respondAndLogError(r.Context(), w, gerrs.ErrUnmarshal)
		return
	}

	user, err = DoBootstrap(r.Context(), user)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	json.NewEncoder(w).Encode(user)
}

// handleHealthz handles "GET /v2/healthz"
func handleHealthz(w http.ResponseWriter, r *http.Request) {
	testUsername, _ := data.HashPassword(time.Now().Local().String())
	testPassword, _ := data.HashPassword(testUsername)
	testUser := rest.User{
		Email:    "healthz@test.user",
		FullName: "Health Check User",
		Username: "healthz" + testUsername[:8],
		Password: testPassword,
	}

	dataAccessLayer, err := dataaccess.Get()
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}

	err = dataAccessLayer.UserCreate(r.Context(), testUser)
	if err != nil {
		log.WithError(err).Warning("health check failure")
		http.Error(w, `{"healthy":false}`, http.StatusServiceUnavailable)
		return
	}
	defer dataAccessLayer.UserDelete(r.Context(), testUser.Username)

	log.Trace("health check pass")
	m := map[string]bool{"healthy": true}
	json.NewEncoder(w).Encode(m)
}

// handleReload handles "GET /v2/reload"
func handleReload(w http.ResponseWriter, r *http.Request) {
	err := config.Reload()
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return
	}
}

func respondAndLogError(ctx context.Context, w http.ResponseWriter, err error) {
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
		log.WithError(err).WithField("status", status).Info(msg)

	// Requested resource doesn't exist
	case gerrs.Is(err, errs.ErrNoSuchBundle):
		fallthrough
	case gerrs.Is(err, errs.ErrNoSuchGroup):
		fallthrough
	case gerrs.Is(err, errs.ErrNoSuchRole):
		fallthrough
	case gerrs.Is(err, errs.ErrNoSuchToken):
		fallthrough
	case gerrs.Is(err, errs.ErrNoSuchUser):
		status = http.StatusNotFound
		log.WithError(err).WithField("status", status).Info(msg)

	// Nope
	case gerrs.Is(err, errs.ErrAdminUndeletable):
		status = http.StatusForbidden
		log.WithError(err).WithField("status", status).Warn(msg)

	// Can't insert over something that already exists
	case gerrs.Is(err, errs.ErrBundleExists):
		fallthrough
	case gerrs.Is(err, errs.ErrGroupExists):
		fallthrough
	case gerrs.Is(err, errs.ErrUserExists):
		status = http.StatusConflict
		log.WithError(err).WithField("status", status).Info(msg)

	// Not done yet
	case gerrs.Is(err, errs.ErrNotImplemented):
		status = http.StatusNotImplemented
		log.WithError(err).WithField("status", status).Info(msg)

	// Data access errors
	case gerrs.Is(err, errs.ErrDataAccessNotInitialized):
		fallthrough
	case gerrs.Is(err, errs.ErrDataAccessCantInitialize):
		fallthrough
	case gerrs.Is(err, errs.ErrDataAccessCantConnect):
		fallthrough
	case gerrs.Is(err, errs.ErrDataAccess):
		status = http.StatusInternalServerError
		log.WithError(err).WithField("status", status).Error(msg)

	// Bad context
	case gerrs.Is(err, gerrs.ErrUnmarshal):
		msg = "Corrupt JSON payload"
		status = http.StatusNotAcceptable
		log.WithError(err).WithField("status", status).Error(msg)

	case gerrs.Is(err, ErrUnauthorized):
		status = http.StatusUnauthorized
		log.WithError(err).WithField("status", status).Error(msg)

	case gerrs.Is(err, ErrGortBundleDisabled):
		status = http.StatusUnauthorized
		if e, ok := err.(gerrs.NestedError); ok {
			err = e.Err
		}
		telemetry.Errors().WithError(err).Commit(ctx)
		log.WithError(err).WithField("status", status).Error(msg)

	// Something else?
	default:
		telemetry.Errors().WithError(err).Commit(ctx)
		status = http.StatusInternalServerError
		log.WithError(err).WithField("status", status).Error("Unhandled server error")
	}

	http.Error(w, msg, status)
}

// Provides a middleware function that simply looks for the EXISTENCE of a valid token.
// More granular role-based auth is also performed at the function level.
func tokenObservingMiddleware(next http.Handler) http.Handler {
	exemptEndpoints := map[string]bool{
		"/v2/authenticate": true,
		"/v2/bootstrap":    true,
		"/v2/healthz":      true,
		"/v2/metrics":      true,
		"/v2/reload":       true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestURI := strings.Split(r.RequestURI, "?")[0]

		if exemptEndpoints[requestURI] {
			next.ServeHTTP(w, r)
			return
		}

		telemetry.TotalRequests().
			WithAttribute("request.uri", r.RequestURI).
			WithAttribute("request.remote-addr", strings.Split(r.RemoteAddr, ":")[0]).
			Commit(r.Context())

		dataAccessLayer, err := dataaccess.Get()
		if err != nil {
			telemetry.Errors().WithError(err).Commit(r.Context())
			return
		}

		token := r.Header.Get("X-Session-Token")
		if token == "" || !dataAccessLayer.TokenEvaluate(r.Context(), token) {
			telemetry.UnauthorizedRequests().
				WithAttribute("request.uri", r.RequestURI).
				WithAttribute("request.remote-addr", strings.Split(r.RemoteAddr, ":")[0]).
				Commit(r.Context())
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func authCommand(handler func(w http.ResponseWriter, r *http.Request), cmd string, subcmd ...string) http.HandlerFunc {
	inner := func(w http.ResponseWriter, r *http.Request) {
		if !authenticateUser(w, r, cmd, subcmd...) {
			return
		}

		handler(w, r)
	}

	return http.HandlerFunc(inner)
}

// authenticateUser is used to authenticate service actions by evaluating them
// against the default Gort command bundle. For example, `authenticateUser(r, "users")`
// is evaluated exactly as if the requesting user executed "gort users" on the
// command line. If the default gort bundle doesn't exist or isn't enabled a
// ErrGortBundleDisabled error will be returned.
// The actual work is done by doAuthenticateUser; this function really cares
// mostly about logging and error handling.
func authenticateUser(w http.ResponseWriter, r *http.Request, gortCommand string, args ...string) bool {
	auth, err := doAuthenticateUser(r, gortCommand, args...)
	if err != nil {
		respondAndLogError(r.Context(), w, err)
		return false
	}

	if !auth {
		respondAndLogError(r.Context(), w, ErrUnauthorized)
	}

	return auth
}

// doAuthenticateUser does the actual work for authenticateUser.
func doAuthenticateUser(r *http.Request, gortCommand string, args ...string) (bool, error) {
	dataAccessLayer, err := dataaccess.Get()
	if err != nil {
		return false, err
	}

	t := r.Header.Get("X-Session-Token")
	if t == "" || !dataAccessLayer.TokenEvaluate(r.Context(), t) {
		return false, ErrUnauthorized
	}

	token, err := dataAccessLayer.TokenRetrieveByToken(r.Context(), t)
	if err != nil {
		return false, err
	}

	perms, err := dataAccessLayer.UserPermissionList(r.Context(), token.User)
	if err != nil {
		return false, err
	}

	bundle, command, err := getGortBundleCommand(r.Context(), gortCommand)
	if err != nil {
		return false, gerrs.Wrap(ErrGortBundleDisabled, err)
	}
	ce := data.CommandEntry{Bundle: bundle, Command: command}

	// Convert all args to types.Value values.
	args = append([]string{gortCommand}, args...)
	argValues, err := types.Inferrer{}.StrictStrings(false).InferAll(args)
	if err != nil {
		return false, err
	}

	env := rules.EvaluationEnvironment{"arg": argValues}

	return auth.EvaluateCommandEntry(perms.Strings(), ce, env)
}

// getGortBundleCommand retrieves the data.BundleCommand value from the default
// Gort command bundle. If the bundle doesn't exist, isn't enabled, or if the
// requested command doesn't exist, an error will be returned.
func getGortBundleCommand(ctx context.Context, commandName string) (data.Bundle, data.BundleCommand, error) {
	const bundleName = "gort"

	dataAccessLayer, err := dataaccess.Get()
	if err != nil {
		return data.Bundle{}, data.BundleCommand{}, err
	}

	bundleVersion, err := dataAccessLayer.BundleEnabledVersion(ctx, bundleName)
	if err != nil {
		return data.Bundle{}, data.BundleCommand{}, err
	}
	if bundleVersion == "" {
		return data.Bundle{}, data.BundleCommand{}, errs.ErrEmptyBundleVersion
	}

	bundle, err := dataAccessLayer.BundleGet(ctx, bundleName, bundleVersion)
	if err != nil {
		return data.Bundle{}, data.BundleCommand{}, err
	}

	cmd, exists := bundle.Commands[commandName]
	if !exists {
		return data.Bundle{}, data.BundleCommand{}, ErrNoSuchCommand
	}

	return bundle, *cmd, nil
}
