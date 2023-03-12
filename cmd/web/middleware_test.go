package main

import (
	"context"
	"github.com/calvarado2004/go-testing-webapp/pkg/data"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test_application_addIPToContext tests the addIPToContext() middleware.
func Test_application_addIPToContext(t *testing.T) {
	tests := []struct {
		headerName  string
		headerValue string
		addr        string
		emptyAddr   bool
	}{
		{"", "", "", false},
		{"", "", "", true},
		{"X-Forwarded-For", "192.10.2.2", "", false},
		{"", "", "hello:world", false},
	}

	// create a dummy handler that will check the context
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// make sure that the value exists in the context
		val := r.Context().Value(contextUserKey)
		if val == nil {
			t.Error("context value is nil")
		}

		// make sure that the value is a string
		ip, ok := val.(string)
		if !ok {
			t.Errorf("context value is not a string")
		}
		t.Log(ip)
	})

	for _, tt := range tests {
		// create the handler to test
		handlerToTest := app.addIPToContext(nextHandler)

		// create a dummy request
		req := httptest.NewRequest("GET", "http://testing", nil)

		if tt.emptyAddr {
			req.RemoteAddr = ""
		}

		if len(tt.headerName) > 0 {
			req.Header.Add(tt.headerName, tt.headerValue)
		}

		if len(tt.addr) > 0 {
			req.RemoteAddr = tt.addr
		}

		// create a dummy response writer
		handlerToTest.ServeHTTP(httptest.NewRecorder(), req)
	}
}

func Test_application_IPFromContext(t *testing.T) {

	// create a context
	ctx := context.Background()

	// put something in the context
	ctx = context.WithValue(ctx, contextUserKey, "192.168.10.2")

	// get the value from the context
	ip := app.ipFromContext(ctx)

	// perform the test
	if !strings.EqualFold(ip, "192.168.10.2") {
		t.Errorf("expected %s but not found", ip)
	}

}

// Test_app_auth tests the auth() middleware.
func Test_app_auth(t *testing.T) {

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})

	var theTests = []struct {
		name   string
		isAuth bool
	}{
		{"logged in", true},
		{"not logged in", false},
	}

	// create a dummy handler that will check the context
	for _, tt := range theTests {
		handlerToTest := app.auth(nextHandler)
		req := httptest.NewRequest("GET", "http://testing", nil)
		req = addContextAndSessionToRequest(req, app)
		if tt.isAuth {
			app.Session.Put(req.Context(), "user", data.User{ID: 1})
		}
		rr := httptest.NewRecorder()
		handlerToTest.ServeHTTP(rr, req)

		if tt.isAuth && rr.Code != http.StatusOK {
			t.Errorf("expected %d but got %d", http.StatusOK, rr.Code)
		}

		if !tt.isAuth && rr.Code != http.StatusTemporaryRedirect {
			t.Errorf("expected %d but got %d", http.StatusTemporaryRedirect, rr.Code)
		}

	}
}
