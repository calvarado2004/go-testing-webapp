package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"net/http"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	// register middleware
	mux.Use(middleware.Recoverer)

	// mux.Use(app.EnableCORS)

	// authentication routes - auth handler, refresh token handler

	// test handler

	// protected routes
	
	return mux
}
