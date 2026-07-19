package main

import (
	"context"
	"embed"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ieee-yp/ecommerce-observability/order-service/internal/catalog"
	"github.com/ieee-yp/ecommerce-observability/order-service/internal/config"
	"github.com/ieee-yp/ecommerce-observability/order-service/internal/db"
	"github.com/ieee-yp/ecommerce-observability/order-service/internal/handler"
	kafkapkg "github.com/ieee-yp/ecommerce-observability/order-service/internal/kafka"
	"github.com/ieee-yp/ecommerce-observability/order-service/internal/logging"
	"github.com/ieee-yp/ecommerce-observability/order-service/internal/repository"
	"github.com/ieee-yp/ecommerce-observability/order-service/internal/router"
	"github.com/ieee-yp/ecommerce-observability/order-service/internal/service"
)

const serviceName = "order-service"

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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	producer := kafkapkg.NewProducer(cfg.KafkaBootstrapServers, logger)
	defer producer.Close()

	repo := repository.NewPostgresOrderRepository(pool)
	svc := service.New(repo, catalog.NewClient(cfg.ProductCatalogURL), producer, logger)

	// Saga consumers run for the lifetime of the process.
	go kafkapkg.RunSagaConsumers(ctx, cfg.KafkaBootstrapServers, svc.HandleSagaEvent, logger)

	r := router.New(
		handler.NewOrderHandler(svc, logger),
		handler.NewHealthHandler(pool, serviceName),
	)

	server := &http.Server{Addr: ":" + cfg.Port, Handler: r}
	go func() {
		logger.Info("starting order-service", "addr", server.Addr)
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
