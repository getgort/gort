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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/dataaccess"
	"github.com/getgort/gort/dataaccess/memory"
)

var adminToken rest.Token

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
	req.Header.Add("X-Session-Token", adminToken.Token)
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
			t.Fatal(err)
		}
	}

	if r.expectedStatus != nil {
		assert.Equal(t, *r.expectedStatus, resp.StatusCode)
	}
}

func createTestRouter() *mux.Router {
	ctx := context.Background()

	// Clear in-memory data for re-running tests
	memory.Reset()

	dataAccessLayer, err := dataaccess.Get()
	if err != nil {
		panic(err)
	}

	userAdmin, err := doBootstrap(ctx, rest.User{})
	if err != nil {
		panic(err)
	}

	adminToken, err = dataAccessLayer.TokenGenerate(ctx, userAdmin.Username, time.Minute)
	if err != nil {
		panic(err)
	}

	router := mux.NewRouter()
	addAllMethodsToRouter(router)

	return router
}
