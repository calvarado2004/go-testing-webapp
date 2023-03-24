package main

import (
	"context"
	"github.com/calvarado2004/go-testing-webapp/pkg/data"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
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

// Test_app_refresh tests the refresh handler function.
func Test_app_refresh(t *testing.T) {

	var theTests = []struct {
		name               string
		token              string
		expectedStatusCode int
		refreshTime        bool
	}{
		{"valid", "", http.StatusOK, true},
		{"valid but not ready", "", http.StatusTooEarly, false},
		{"invalid", "invalid", http.StatusUnauthorized, false},
		{"expired", expiredToken, http.StatusUnauthorized, false},
	}

	testUser := data.User{
		ID:        2,
		FirstName: "Test",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	oldRefreshTime := refreshTokenExpiry

	for _, tt := range theTests {
		var tkn string
		if tt.token == "" {
			if tt.refreshTime {
				refreshTokenExpiry = time.Second * 1
			}

			tokens, _ := app.generateTokenPair(&testUser)
			tkn = tokens.RefreshToken
		} else {
			tkn = tt.token
		}

		postedData := url.Values{
			"refresh_token": {tkn},
		}

		req, err := http.NewRequest("POST", "/v1/refresh-token", strings.NewReader(postedData.Encode()))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(app.refresh)

		handler.ServeHTTP(rr, req)

		if tt.expectedStatusCode != rr.Code {
			t.Errorf("handler returned wrong status code: got %v want %v, test %s", rr.Code, tt.expectedStatusCode, tt.name)
		}

		refreshTokenExpiry = oldRefreshTime

	}

}

// Test_app_userHandlers tests the user handlers.
func Test_app_userHandlers(t *testing.T) {
	var theTests = []struct {
		name           string
		method         string
		json           string
		paramID        string
		handler        http.HandlerFunc
		expectedStatus int
	}{
		{"allUsers", "GET", "", "", app.allUsers, http.StatusOK},
		{"deleteUser", "DELETE", "", "1", app.deleteUser, http.StatusNoContent},
		{"getUser valid", "GET", "", "1", app.getUser, http.StatusOK},
	}

	for _, tt := range theTests {
		var req *http.Request
		if tt.json != "" {
			req, _ = http.NewRequest(tt.method, "/v1/users", strings.NewReader(tt.json))
		} else {
			req, _ = http.NewRequest(tt.method, "/v1/users", nil)
		}
		if tt.paramID != "" {
			chiCtx := chi.NewRouteContext()

			chiCtx.URLParams.Add("userID", tt.paramID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
			rr := httptest.NewRecorder()

			tt.handler.ServeHTTP(rr, req)

			if tt.expectedStatus != rr.Code {
				t.Errorf("handler returned wrong status code: got %v want %v, test %s", rr.Code, tt.expectedStatus, tt.name)
			}

		}
	}
}
