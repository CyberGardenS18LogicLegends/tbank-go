package main

import (
	"log/slog"
	"net/http"
	"os"
	"tbank-go/internal/config"
	"tbank-go/internal/services/auth"
	"tbank-go/internal/services/expenses"
	"tbank-go/internal/services/geminiAnalysis"
	"tbank-go/internal/services/incomes"
	"tbank-go/internal/services/users"
	"tbank-go/internal/sqlite"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	httpSwagger "github.com/swaggo/http-swagger/v2"
	_ "tbank-go/docs"
)

const (
	envDev  = "dev"
	envProd = "prod"
)

// @title TBank API
// @version 1.4
// @description Api CG T-Bank Finance management

// @host https://cg-api.ffokildam.ru:8443/
// @BasePath /api
func main() {

	//certFile := "C:\\Certbot\\live\\cg-api.ffokildam.ru\\fullchain.pem"
	//	keyFile := "C:\\Certbot\\live\\cg-api.ffokildam.ru\\privkey.pem"
	cfg := config.MustLoadConfig()
	log := setupLogger(cfg.Env)

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

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:63342"}, // Allow specific origin
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true, // Allow cookies, authorization headers, etc.
		MaxAge:           300,  // Cache preflight requests for 5 minutes
	}))

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/swagger/*", httpSwagger.WrapHandler)

	router.Route("/api", func(r chi.Router) {
		r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
			auth.Register(db, w, r, log)
		})
		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			auth.Login(db, w, r, log, cfg.JwtSecret, cfg.JwtLifetime)
		})
		r.Post("/change-password", func(w http.ResponseWriter, r *http.Request) {
			auth.ChangePassword(db, w, r, log)
		})
		r.With(auth.AuthMiddleware(cfg.JwtSecret, log)).Route("/income", func(r chi.Router) {
			r.Post("/", incomes.AddIncomeHandler(db, log))
			r.Get("/", incomes.GetIncomesHandler(db, log))
			r.Delete("/{id}", incomes.DeleteIncomeHandler(db, log))
		})
		r.With(auth.AuthMiddleware(cfg.JwtSecret, log)).Route("/expense", func(r chi.Router) {
			r.Post("/", expenses.AddExpenseHandler(db, log))
			r.Get("/", expenses.GetExpensesHandler(db, log))
			r.Delete("/{id}", expenses.DeleteExpenseHandler(db, log))
		})
		r.With(auth.AuthMiddleware(cfg.JwtSecret, log)).Route("/users", func(r chi.Router) {
			r.Put("/", users.UpdateUserNamesHandler(db, log))
			r.Get("/", users.GetUserInfoHandler(db, log))
		})
		r.With(auth.AuthMiddleware(cfg.JwtSecret, log)).Route("/ai-advice", func(r chi.Router) {
			r.Get("/", geminiAnalysis.GenerateFinancialAdviceHandler(db, log))
		})
	})

	// Start the HTTPS server
	log.Info("starting HTTPS server", slog.String("address", cfg.Address))
	if err := http.ListenAndServeTLS(cfg.Address, "server.crt", "server.key", router); err != nil {
		log.Error("HTTPS server failed", slog.Any("error", err))
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
