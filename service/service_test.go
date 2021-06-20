package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/getgort/gort/dataaccess/memory"
)

func jsonRequest(t *testing.T, router *mux.Router, method string, target string, body interface{}, out interface{}) int {
	var bodyReader io.Reader = nil
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}
	req := httptest.NewRequest(method, target, bodyReader)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if out != nil {
		err = json.Unmarshal(resBody, &out)
		if err != nil {
			fmt.Println(string(resBody))
			t.Fatal(err)
		}
	}
	return resp.StatusCode
}

func createTestRouter() *mux.Router {
	dataAccessLayer = memory.NewInMemoryDataAccess()
	router := mux.NewRouter()
	addAllMethodsToRouter(router)
	return router
}
