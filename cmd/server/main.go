package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gistlike/internal/config"
	"gistlike/internal/handler"
	"gistlike/internal/repository"
	"gistlike/internal/service"

	_ "modernc.org/sqlite"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if err := os.MkdirAll(filepath.Dir(cfg.DatabasePath), 0o755); err != nil {
		logger.Error("failed to prepare data directory", "error", err)
		os.Exit(1)
	}

	db, err := sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// SQLite works best with a single writer connection for simple MVP services.
	db.SetMaxOpenConns(1)
	db.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logger.Error("failed to connect database", "error", err)
		os.Exit(1)
	}

	repo := repository.NewSQLiteSnippetRepository(db)
	if err := repo.EnsureSchema(ctx); err != nil {
		logger.Error("failed to initialize schema", "error", err)
		os.Exit(1)
	}

	snippetService := service.NewSnippetService(repo)
	router, err := handler.SetupRouter(cfg, snippetService)
	if err != nil {
		logger.Error("failed to setup router", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              cfg.Address,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("starting gistlike server", "addr", cfg.Address, "db", cfg.DatabasePath)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}
