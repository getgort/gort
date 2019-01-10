package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/clockworksoul/cog2/dal"
	"github.com/clockworksoul/cog2/dal/memory"
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

			requestLine := fmt.Sprintf("%s %s %s",
				r.Method,
				r.RequestURI,
				r.Proto)

			e := RequestEvent{
				Addr:      r.RemoteAddr,
				UserID:    "-", // TODO Identify logged in users.
				Timestamp: time.Now(),
				Request:   requestLine,
				Status:    status,
				Size:      int64(bytelen),
			}

			logsous <- e
		})
	}
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
	router.Use(buildLoggingMiddleware(requests))
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
			dataAccessLayer = memory.NewInMemoryDataAccess()
			err := dataAccessLayer.Initialize()

			if err != nil {
				log.Warn("[InitializeDataAccessLayer] Failed to connect to data source: ", err.Error())
				log.Infof("[InitializeDataAccessLayer] Waiting %d seconds to try again", delay)

				<-time.After(delay * time.Second)

				delay *= 2

				if delay > 60 {
					delay = 60
				}
			}

			dataAccessLayerInitialized = true
		}

		log.Info("[InitializeDataAccessLayer] Connection to data source established")
	}()
}
