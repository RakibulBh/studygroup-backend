package main

import (
	"net/http"
	"time"

	"github.com/RakibulBh/studygroup-backend/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type application struct {
	config config
	store  store.Storage
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type config struct {
	addr   string
	db     dbConfig
	auth   auth
	env    string
	apiURL string
}

type auth struct {
	jwtSecret  string
	exp        time.Duration
	refreshExp time.Duration
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Healthcheck
	r.Route("/health", func(r chi.Router) {
		r.Get("/", app.Healthcheck)
	})

	r.Route("/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", app.Register)
			r.Post("/login", app.Login)
			// r.Post("/logout", app.Logout)
			r.Get("/refresh", app.Refresh)
		})

		r.Route("/groups", func(r chi.Router) {
			r.Use(app.Authenticate)
			r.Get("/", app.GetUserGroups)
			r.Get("/{id}", app.GetGroup)
			r.Post("/", app.CreateGroup)
			r.Put("/{id}", app.UpdateGroup)
		})

		r.Route("/sessions", func(r chi.Router) {
			r.Get("/", app.GetSessions)
			r.Get("/{id}", app.GetSession)
		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}

	return srv.ListenAndServe()
}
