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
	mux.Post("/v1/auth", app.authenticate)
	mux.Post("/v1/refresh", app.refresh)

	// test handler

	// protected routes
	mux.Route("/v1/users", func(muxAuth chi.Router) {

		muxAuth.Get("/", app.allUsers)
		muxAuth.Get("/{id}", app.getUser)
		muxAuth.Patch("/{id}", app.updateUser)
		muxAuth.Delete("/{id}", app.deleteUser)
		muxAuth.Post("/", app.insertUser)

	})

	return mux
}
