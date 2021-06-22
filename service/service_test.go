package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/getgort/gort/dataaccess/memory"
)

// ResponseTester allows a single HTTP test to be sent to a router and the
// result to be verified.
type ResponseTester struct {
	body           interface{}
	out            interface{}
	method         string
	target         string
	expectedStatus *int
}

// NewResponseTester creates a ResponseTester for the given method and target URL.
func NewResponseTester(method, target string) ResponseTester {
	return ResponseTester{
		method: method,
		target: target,
	}
}

// WithBody adds a body to a request to be tested.
// The input value is JSON encoded.
func (r ResponseTester) WithBody(body interface{}) ResponseTester {
	r.body = body
	return r
}

// WithStatus adds an expected HTTP status to a ResponseTester.
// If the response does not have the specified code, the test fails.
func (r ResponseTester) WithStatus(status int) ResponseTester {
	r.expectedStatus = &status
	return r
}

// WithOutput requests JSON output from a response.
// The provided pointer will be populated with the unmarshaled form of the JSON
// in the response body.
func (r ResponseTester) WithOutput(out interface{}) ResponseTester {
	r.out = out
	return r
}

// Test sends a constructed request to the provided router and performs any
// required checks on it.
func (r ResponseTester) Test(t *testing.T, router *mux.Router) {
	t.Helper()

	var bodyReader io.Reader = nil
	if r.body != nil {
		jsonData, err := json.Marshal(r.body)
		if err != nil {
			t.Fatal(err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req := httptest.NewRequest(r.method, r.target, bodyReader)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if r.out != nil {
		err = json.Unmarshal(resBody, &r.out)
		if err != nil {
			fmt.Println(string(resBody))
			t.Fatal(err)
		}
	}

	if r.expectedStatus != nil {
		assert.Equal(t, *r.expectedStatus, resp.StatusCode)
	}
}

func createTestRouter() *mux.Router {
	dataAccessLayer = memory.NewInMemoryDataAccess()
	router := mux.NewRouter()
	addAllMethodsToRouter(router)
	return router
}
