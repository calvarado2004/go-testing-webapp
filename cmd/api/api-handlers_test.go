package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test_app_authenticate tests the authenticate handler function.
func Test_app_authenticate(t *testing.T) {
	var theTests = []struct {
		name               string
		requestBody        string
		expectedStatusCode int
	}{
		{"valid", `{"email":"admin@example.com","password":"secret"}`, http.StatusOK},
		{"invalidJSON", `This is not json`, http.StatusUnauthorized},
		{"emptyJSON", `{}`, http.StatusUnauthorized},
		{"emptyEmail", `{"email":"","password":"secret"}`, http.StatusUnauthorized},
		{"emptyPassword", `{"email":"admin@example.com","password":""}`, http.StatusUnauthorized},
		{"invalidUser", `{"email":"admin@otherdomain.com","password":"secret"}`, http.StatusUnauthorized},
	}

	for _, tt := range theTests {
		var reader io.Reader
		reader = strings.NewReader(tt.requestBody)

		req, err := http.NewRequest("POST", "/v1/auth", reader)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(app.authenticate)

		handler.ServeHTTP(rr, req)

		if tt.expectedStatusCode != rr.Code {
			t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.expectedStatusCode)
		}

	}
}