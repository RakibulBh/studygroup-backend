package main

import (
	"log"
	"time"

	"github.com/RakibulBh/studygroup-backend/internal/db"
	"github.com/RakibulBh/studygroup-backend/internal/env"
	"github.com/RakibulBh/studygroup-backend/internal/store"
)

func main() {

	cfg := config{
		env:    env.GetString("ENV", "development"),
		addr:   ":" + env.GetString("PORT", "8080"),
		apiURL: env.GetString("API_URL", "http://localhost:8080"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost:5432/shaheed?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 10),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 10),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "10s"),
		},
		auth: auth{
			jwtSecret:  env.GetString("AUTH_SECRET", "VERYSECRET"),
			exp:        env.GetDuration("AUTH_EXP", time.Hour*200),
			refreshExp: env.GetDuration("AUTH_REFRESH_EXP", time.Hour*24*7), // 7 days
		},
		llm: llmConfig{
			model:  "gemini-2.0-flash-lite",
			apiKey: env.GetString("GEMINI_API_KEY", "API_KEY_HERE"),
		},
	}

	// Database
	db, err := db.New(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Store
	store := store.NewStorage(db)

	app := &application{
		config: cfg,
		store:  store,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
