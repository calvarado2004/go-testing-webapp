package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"strings"
	"testing"
)

// Test_application_routes tests that the routes are registered with chi
func Test_application_routes(t *testing.T) {

	var registeredRoutes = []struct {
		route  string
		method string
	}{
		{"/", "GET"},
		{"/login", "POST"},
		{"/user/login", "GET"},
		{"/static/*", "GET"},
	}

	mux := app.routes()

	chiRoutes := mux.(chi.Routes)

	for _, route := range registeredRoutes {

		if !routeExists(route.route, route.method, chiRoutes) {
			t.Errorf("Expected route %s with method %s to be registered", route.route, route.method)
		}

	}
}

// routeExists checks if a route is registered with chi
func routeExists(testRoute string, testMethod string, chiRoutes chi.Routes) bool {

	found := false

	_ = chi.Walk(chiRoutes, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if strings.EqualFold(method, testMethod) && strings.EqualFold(route, testRoute) {
			found = true
		}
		return nil
	})

	return found
}
