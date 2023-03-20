package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test_app_enableCORS tests the enableCORS middleware
func Test_app_enableCORS(t *testing.T) {

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	var theTests = []struct {
		name           string
		method         string
		expectedHeader bool
	}{
		{"preflight", "OPTIONS", true},
		{"get", "GET", false},
		{"post", "POST", false},
	}

	for _, tt := range theTests {
		handlerToTest := app.enableCORS(nextHandler)
		req := httptest.NewRequest(tt.method, "http://testing", nil)
		rr := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rr, req)

		if tt.expectedHeader && rr.Header().Get("Access-Control-Allow-Credentials") == "" {
			t.Errorf("Expected header not found on test %s", tt.name)
		}

		if !tt.expectedHeader && rr.Header().Get("Access-Control-Allow-Credentials") != "" {
			t.Errorf("Unexpected header found on test %s", tt.name)
		}

	}
}