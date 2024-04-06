package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"catalog/internal/config"
	"catalog/internal/http-handlers/catalog"
	delete "catalog/internal/http-handlers/delete"
	edit "catalog/internal/http-handlers/edit"
	"catalog/internal/http-handlers/new"
	"catalog/internal/lib/logger/sl"
	postgres "catalog/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

const (
	cfgPath = "C:/Users/Leonid/Desktop/catalog/config/config.env"
)

func main() {
	// Loads config values from .env into the system
	if err := godotenv.Load(cfgPath); err != nil {
		log.Fatal("No .env file found")
	}
	cfg := config.MustLoad()

	// Init logger
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	storage, err := postgres.New(cfg.SQLDriver, cfg.SQLConnectionInfo)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	defer storage.DB.Close()
	log.Info("connected to PostgreSQL server")

	if err := storage.UpMigration(cfg.SQLMigrationInfo); err != nil {
		log.Debug("failed to migrate DB", sl.Err(err))
	}

	// Use chi and help middleware
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Get("/catalog", catalog.New(log, storage))
	router.Post("/new", new.New(log, storage))
	router.Post("/delete", delete.New(log, storage))
	router.Post("/edit", edit.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.HTTPServerAddress))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:    cfg.HTTPServerAddress,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")

	// Ending all contexts
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server")

		return
	}

	log.Info("server stopped")
}
