package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL           string
	Port                  string
	KafkaBootstrapServers []string
	ProductCatalogURL     string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		Port:              os.Getenv("PORT"),
		ProductCatalogURL: strings.TrimRight(os.Getenv("PRODUCT_CATALOG_URL"), "/"),
	}
	if brokers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS"); brokers != "" {
		cfg.KafkaBootstrapServers = strings.Split(brokers, ",")
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.Port == "" {
		return nil, fmt.Errorf("PORT is required")
	}
	if len(cfg.KafkaBootstrapServers) == 0 {
		return nil, fmt.Errorf("KAFKA_BOOTSTRAP_SERVERS is required")
	}
	if cfg.ProductCatalogURL == "" {
		return nil, fmt.Errorf("PRODUCT_CATALOG_URL is required")
	}
	return cfg, nil
}
