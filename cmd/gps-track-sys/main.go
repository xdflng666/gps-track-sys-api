package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gps-track-sys-api/internal/config"
	mwLogger "gps-track-sys-api/internal/http-server/middleware/logger"
	"gps-track-sys-api/internal/lib/logger/handlers/slogpretty"
	"gps-track-sys-api/internal/lib/logger/sl"
	"gps-track-sys-api/internal/storage/sqlite"
	"log"
	"log/slog"
	"net/http"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	err := os.Setenv("CONFIG_PATH", "./config/local.yaml")
	if err != nil {
		log.Fatal("Can't set CONFIG_PATH")
	} // Local config path

	// Init config : cleanenv
	cfg := config.MustLoad()
	fmt.Println(cfg)

	// Init logger : slog
	logger := setupLogger(cfg.Env)

	logger.Info("Starting application", slog.String("env", cfg.Env))
	logger.Debug("Debug messages enabled")

	// Init storage : sqlite
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		logger.Error("Failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	_ = storage // kostyl

	// Init router : chi, chi render
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.RequestID) // used for tracing
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(logger))
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer) // handles panic
	router.Use(middleware.URLFormat)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	router.Get("/bye", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Goodbye World!"))
	})

	// Run server

	logger.Info("Starting server at ", slog.String("addr", cfg.Address))

	srv := &http.Server{
		Handler:      router,
		Addr:         cfg.HTTPServer.Address,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		logger.Error("Failed to start HTTP Server")
	}

	logger.Error("Server stopped unexpectedly")

}

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case envLocal:
		logger = setupPrettySlog()
	case envDev:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return logger
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
