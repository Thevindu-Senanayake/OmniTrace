package main

import (
	"context"
	"embed"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ieee-yp/ecommerce-observability/product-catalog/internal/config"
	"github.com/ieee-yp/ecommerce-observability/product-catalog/internal/db"
	"github.com/ieee-yp/ecommerce-observability/product-catalog/internal/handler"
	"github.com/ieee-yp/ecommerce-observability/product-catalog/internal/logging"
	"github.com/ieee-yp/ecommerce-observability/product-catalog/internal/observability"
	"github.com/ieee-yp/ecommerce-observability/product-catalog/internal/repository"
	"github.com/ieee-yp/ecommerce-observability/product-catalog/internal/router"
)

const serviceName = "product-catalog"

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Observability first: the otelslog bridge inside logging.New binds to the
	// global LoggerProvider at construction, so the providers must exist before
	// the logger is built. The exporters dial lazily, so this never blocks on
	// the collector being up.
	shutdownObs, err := observability.Setup(ctx, serviceName)
	if err != nil {
		slog.Error("observability setup failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdownObs(shutdownCtx)
	}()

	logger := logging.New(serviceName)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	if err := db.RunMigrations(migrationsFS, "migrations", cfg.DatabaseURL); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}
	logger.Info("migrations applied")

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.NewPostgresProductRepository(pool)
	r := router.New(
		handler.NewProductHandler(repo, logger),
		handler.NewHealthHandler(pool, serviceName),
	)

	server := &http.Server{Addr: ":" + cfg.Port, Handler: r}
	go func() {
		logger.Info("starting product-catalog", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server exited", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}
