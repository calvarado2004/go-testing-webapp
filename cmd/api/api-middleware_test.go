package main

import (
	"fmt"
	"github.com/calvarado2004/go-testing-webapp/pkg/data"
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

// Test_app_authRequired tests the authRequired middleware
func Test_app_authRequired(t *testing.T) {

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	testUser := data.User{
		ID:        1,
		FirstName: "Test",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	tokens, _ := app.generateTokenPair(&testUser)

	var theTests = []struct {
		name               string
		token              string
		expectedAuthorized bool
		setHeader          bool
	}{
		{name: "valid token", token: fmt.Sprintf("Bearer %s", tokens.AccessToken), expectedAuthorized: true, setHeader: true},
		{name: "no token", token: fmt.Sprintf("Bearer "), expectedAuthorized: false, setHeader: false},
		{name: "invalid token", token: fmt.Sprintf("Bearer %s", expiredToken), expectedAuthorized: false, setHeader: true},
	}

	for _, tt := range theTests {
		req, _ := http.NewRequest("GET", "/", nil)
		if tt.setHeader {
			req.Header.Set("Authorization", tt.token)
		}

		rr := httptest.NewRecorder()

		handlerToTest := app.authRequired(nextHandler)

		handlerToTest.ServeHTTP(rr, req)

		if tt.expectedAuthorized && rr.Code != http.StatusOK {
			t.Errorf("Expected authorized request to return 200 on test %s", tt.name)
		}

		if !tt.expectedAuthorized && rr.Code != 400 {
			t.Errorf("Expected unauthorized request to return 401 on test %s, returned %d", tt.name, rr.Code)
		}

	}

}
