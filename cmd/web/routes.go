package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func (app *application) routes() http.Handler {

	mux := chi.NewRouter()

	// register middleware for unauthenticated routes
	mux.Use(middleware.Recoverer)
	mux.Use(app.addIPToContext)
	mux.Use(app.Session.LoadAndSave)
	// register the unauthenticated routes
	mux.Get("/", app.Home)
	mux.Post("/login", app.Login)
	mux.Get("/user/login", app.Login)

	// register middleware for authenticated routes
	mux.Route("/user/profile", func(muxAuth chi.Router) {
		muxAuth.Use(app.auth)
		muxAuth.Get("/", app.Profile)
		muxAuth.Post("/upload-profile-pic", app.UploadProfilePic)
	})

	// static files
	fileServer := http.FileServer(http.Dir("./static/"))

	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// mux satisfies the http.Handler interface
	return mux
}
