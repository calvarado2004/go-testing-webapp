package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

type contextKey string

const contextUserKey contextKey = "user_ip"

func (app *application) ipFromContext(ctx context.Context) string {

	return ctx.Value(contextUserKey).(string)
}

func (app *application) addIPToContext(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		// Get the IP address from the request and add it to the context.
		ip, err := getIP(r)
		if err != nil {
			ip, _, _ = net.SplitHostPort(r.RemoteAddr)
			if len(ip) == 0 {
				ip = "unknown"
			}

			ctx = context.WithValue(r.Context(), contextUserKey, ip)

		} else {
			ctx = context.WithValue(r.Context(), contextUserKey, ip)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getIP(r *http.Request) (string, error) {

	//192.0.0.1:1234
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "unknown", err
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		return "unknown", fmt.Errorf("invalid IP address: %s", r.RemoteAddr)
	}

	forward := r.Header.Get("X-Forwarded-For")
	if len(forward) > 0 {

		ip = forward
	}

	if len(ip) == 0 {

		ip = "forwarded"

	}

	return ip, nil
}

// auth is a middleware that checks if a user is authenticated (signed in).
func (app *application) auth(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.Session.Exists(r.Context(), "user") {
			app.Session.Put(r.Context(), "error", "You must be logged in to access that page")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}
