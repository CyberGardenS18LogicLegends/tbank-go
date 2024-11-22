package main

import (
	"log/slog"
	"net/http"
	"os"
	"tbank-go/internal/config"
	"tbank-go/internal/services"
	"tbank-go/internal/sqlite"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envDev  = "dev"
	envProd = "prod"
)

func main() {
	cfg := config.MustLoadConfig()
	log := setupLogger(cfg.Env)

	// Initialize the database
	db, err := sqlite.InitializeDatabase(cfg.StoragePath, log)
	if err != nil {
		log.Error("failed to initialize database", slog.Any("error", err))
		os.Exit(1)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Error("failed to close database", slog.Any("error", err))
		}
	}()

	log.Info("config loaded", slog.String("env", cfg.Env))
	log.Debug("debug messages enabled")

	// Set up router
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Define API routes
	router.Route("/api", func(r chi.Router) {
		r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
			services.Register(db, w, r, log)
		})

		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			services.Login(db, w, r, log, cfg.JwtSecret)
		})

		r.Post("/change-password", func(w http.ResponseWriter, r *http.Request) {
			services.ChangePassword(db, w, r, log)
		})
	})

	// Start the server
	log.Info("starting server", slog.String("address", cfg.Address))
	if err := http.ListenAndServe(cfg.Address, router); err != nil {
		log.Error("server failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
