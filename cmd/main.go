package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mwLogger "github.com/Noviiich/io-bound-task/internal/http-server/middleware/logger"

	"github.com/Noviiich/io-bound-task/internal/config"
	"github.com/Noviiich/io-bound-task/internal/http-server/handlers/tasks/create"
	"github.com/Noviiich/io-bound-task/internal/http-server/handlers/tasks/delete"
	"github.com/Noviiich/io-bound-task/internal/http-server/handlers/tasks/get"
	"github.com/Noviiich/io-bound-task/internal/lib/logger"
	"github.com/Noviiich/io-bound-task/internal/lib/logger/sl"
	tServices "github.com/Noviiich/io-bound-task/internal/services"
	"github.com/Noviiich/io-bound-task/internal/storage/memory"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.MustLoad()

	log := logger.New(cfg.Env)

	log.Info(
		"starting task service",
		slog.String("env", cfg.Env),
		slog.String("version", "v0.0.1"),
	)

	storage := memory.New()
	log.Info("starting memory storage")

	services := tServices.New(storage)
	log.Info("task service initialized", slog.String("storage", "memory"))

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/tasks", func(r chi.Router) {
		r.Post("/", create.New(log, services))
		r.Get("/{id}", get.New(log, services))
		// r.Put("/{id}", update.New(log, services))
		r.Delete("/{id}", delete.New(log, services))
	})

	log.Info("starting server", slog.String("address", cfg.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error(err.Error())
		}
	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	log.Info("server stopped")
}
