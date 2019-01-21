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
			if token != "" {
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
		"/authenticate": true,
		"/healthz":      true,
		"/metrics":      true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestURI := strings.Split(r.RequestURI, "?")[0]

		if !exemptEndpoints[requestURI] {
			next.ServeHTTP(w, r)
		}

		token := r.Header.Get("X-Session-Token")
		if token == "" || !dataAccessLayer.TokenEvaluate(token)
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

// handleHealthz handles "GET /authenticate?username={username}&password={password}}"
// TODO Can we make this more meaningful?
func handleAuthenticate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	password := vars["password"]

	authenticated, err := dataAccessLayer.UserAuthenticate(username, password)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !authenticated {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	token, err := dataAccessLayer.TokenRetrieveByUser(username)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(token)
}

// handleHealthz handles "GET /healthz}"
// TODO Can we make this more meaningful?
func handleHealthz(w http.ResponseWriter, r *http.Request) {
	m := map[string]bool{"healthy": true}

	json.NewEncoder(w).Encode(m)
}

func addHealthzMethodToRouter(router *mux.Router) {
	router.HandleFunc("/authenticate", handleAuthenticate).
		Methods("GET").
		Queries("username", "{username}", "password", "{password}")

	router.HandleFunc("/healthz", handleHealthz).
		Methods("GET")
}
