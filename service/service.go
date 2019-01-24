package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/clockworksoul/cog2/config"
	"github.com/clockworksoul/cog2/dal"
	"github.com/clockworksoul/cog2/dal/postgres"
	"github.com/clockworksoul/cog2/data/rest"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var (
	dataAccessLayerInitialized bool
	dataAccessLayer            dal.DataAccess
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
	// If user doesn't have a defined email, we default to "cog@localhost".
	if user.Email == "" {
		user.Email = "cog@localhost"
	}

	// If user doesn't have a defined name, we default to "Cog Administrator".
	if user.FullName == "" {
		user.FullName = "Cog Administrator"
	}

	// If user doesn't have a defined email, we default to "admin".
	if user.Username == "" {
		user.Username = "admin"
	}

	// If user doesn't have a defined password, we kindly generate one.
	if user.Password == "" {
		password, err := dal.GenerateRandomToken(32)
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

// RESTServer represents a Cog REST API service.
type RESTServer struct {
	*http.Server

	requests chan RequestEvent
}

// BuildRESTServer builds a RESTServer.
func BuildRESTServer(addr string) *RESTServer {
	InitializeDataAccessLayer()

	requests := make(chan RequestEvent)

	router := mux.NewRouter()
	router.Use(buildLoggingMiddleware(requests), tokenObservingMiddleware)

	addHealthzMethodToRouter(router)
	addUserMethodsToRouter(router)
	addGroupMethodsToRouter(router)

	server := &http.Server{Addr: addr, Handler: router}

	return &RESTServer{server, requests}
}

// Requests retrieves the channel to which user request events are sent.
func (s *RESTServer) Requests() <-chan RequestEvent {
	return s.requests
}

// ListenAndServe starts the Cog web service.
func (s *RESTServer) ListenAndServe() error {
	log.Printf("[RESTServer.ListenAndServe] Cog service is starting on " + s.Addr)

	return s.Server.ListenAndServe()
}

// InitializeDataAccessLayer will initialize the data access layer, if it
// isn't already initialized. It is called automatically by BuildRESTServer().
func InitializeDataAccessLayer() {
	go func() {
		var delay time.Duration = 1

		for !dataAccessLayerInitialized {
			dataAccessLayer = postgres.NewPostgresDataAccess(config.GetDatabaseConfigs())
			err := dataAccessLayer.Initialize()

			if err != nil {
				log.Warn("[InitializeDataAccessLayer] Failed to connect to data source: ", err.Error())
				log.Infof("[InitializeDataAccessLayer] Waiting %d seconds to try again", delay)

				<-time.After(delay * time.Second)

				delay *= 2

				if delay > 60 {
					delay = 60
				}
			} else {
				dataAccessLayerInitialized = true
			}
		}

		log.Info("[InitializeDataAccessLayer] Connection to data source established")
	}()
}

// handleAuthenticate handles "GET /authenticate"
func handleAuthenticate(w http.ResponseWriter, r *http.Request) {
	// Grab the user struct from the request. If it doesn't exist, respond with
	// a client error.
	user := rest.User{}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Missing user data", http.StatusBadRequest)
		log.Errorf("[handleAuthenticate.1] %s", "Missing user data")
		return
	}

	username := user.Username
	password := user.Password

	authenticated, err := dataAccessLayer.UserAuthenticate(username, password)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Errorf("[handleAuthenticate.2] %s", err.Error())
		return
	}

	if !authenticated {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	token, err := dataAccessLayer.TokenGenerate(username, 10*time.Minute)
	if err != nil {
		respondAndLogServerError(w, err, "handleAuthenticate", 3)
		return
	}

	json.NewEncoder(w).Encode(token)
}

func respondAndLogServerError(w http.ResponseWriter, err error, label string, index int) {
	http.Error(w, "Internal server error", http.StatusInternalServerError)
	log.Errorf("[%s.%d] %s", label, index, err.Error())
}

// handleBootstrap handles "POST /bootstrap"
func handleBootstrap(w http.ResponseWriter, r *http.Request) {
	users, err := dataAccessLayer.UserList()
	if err != nil {
		respondAndLogServerError(w, err, "handleBootstrap", 1)
		return
	}

	// If we already have users on this host, reject as "already bootstrapped".
	if len(users) != 0 {
		http.Error(w, "Service already bootstrapped", http.StatusConflict)
		log.Warn("[handleBootstrap.2] Re-bootstrap attempted")
		return
	}

	// Grab the user struct from the request. If it doesn't exist, respond with
	// a client error.
	user := rest.User{}
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Missing user data", http.StatusBadRequest)
		log.Errorf("[handleBootstrap.3] %s", "Missing user data")
		return
	}

	// Set user defaults where necessary.
	user, err = bootstrapUserWithDefaults(user)
	if err != nil {
		respondAndLogServerError(w, err, "handleBootstrap", 4)
		return
	}

	// Persist our shiny new user to the database.
	err = dataAccessLayer.UserCreate(user)
	if err != nil {
		respondAndLogServerError(w, err, "handleBootstrap", 5)
		return
	}

	// Create cog-admin group. This currently can't be customized.
	group := rest.Group{Name: "cog-admin"}
	err = dataAccessLayer.GroupCreate(group)
	if err != nil {
		respondAndLogServerError(w, err, "handleBootstrap", 6)
		return
	}

	// Add the admin user to the cog-admin group
	err = dataAccessLayer.GroupAddUser(group.Name, user.Username)
	if err != nil {
		respondAndLogServerError(w, err, "handleBootstrap", 7)
		return
	}

	json.NewEncoder(w).Encode(user)
}

// handleHealthz handles "GET /healthz}"
// TODO Can we make this more meaningful?
func handleHealthz(w http.ResponseWriter, r *http.Request) {
	m := map[string]bool{"healthy": true}

	json.NewEncoder(w).Encode(m)
}

func addHealthzMethodToRouter(router *mux.Router) {
	router.HandleFunc("/v2/authenticate", handleAuthenticate).
		Methods("POST")

	router.HandleFunc("/v2/bootstrap", handleBootstrap).
		Methods("POST")

	router.HandleFunc("/v2/healthz", handleHealthz).
		Methods("GET")
}
