package main

import (
	"context"
	"embed"
	"net/http"
	"os"

	"github.com/ieee-yp/ecommerce-observability/product-catalog/internal/config"
	"github.com/ieee-yp/ecommerce-observability/product-catalog/internal/db"
	"github.com/ieee-yp/ecommerce-observability/product-catalog/internal/handler"
	"github.com/ieee-yp/ecommerce-observability/product-catalog/internal/logging"
	"github.com/ieee-yp/ecommerce-observability/product-catalog/internal/repository"
	"github.com/ieee-yp/ecommerce-observability/product-catalog/internal/router"
)

const serviceName = "product-catalog"

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
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

	pool, err := db.NewPool(context.Background(), cfg.DatabaseURL)
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

	addr := ":" + cfg.Port
	logger.Info("starting product-catalog", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Error("server exited", "error", err)
		os.Exit(1)
	}
}
