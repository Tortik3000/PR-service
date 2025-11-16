package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
)

type (
	Config struct {
		REST
		PG
		Observability
	}

	REST struct {
		Port string `setEnv:"PORT"`
	}

	PG struct {
		URL      string
		Host     string `setEnv:"POSTGRES_HOST"`
		Port     string `setEnv:"POSTGRES_PORT"`
		DB       string `setEnv:"POSTGRES_DB"`
		User     string `setEnv:"POSTGRES_USER"`
		Password string `setEnv:"POSTGRES_PASSWORD"`
	}

	Observability struct {
		MetricsPort string `env:"METRICS_PORT"`
	}
)

func New() (*Config, error) {
	cfg := &Config{}

	envVars := map[string]*string{
		"REST_PORT":         &cfg.REST.Port,
		"POSTGRES_HOST":     &cfg.PG.Host,
		"POSTGRES_PORT":     &cfg.PG.Port,
		"POSTGRES_DB":       &cfg.PG.DB,
		"POSTGRES_USER":     &cfg.PG.User,
		"POSTGRES_PASSWORD": &cfg.PG.Password,
		"METRICS_PORT":      &cfg.Observability.MetricsPort,
	}

	for name, ptr := range envVars {
		if err := requireEnv(ptr, name); err != nil {
			return nil, err
		}
	}

	cfg.PG.URL = fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		url.QueryEscape(cfg.PG.User),
		url.QueryEscape(cfg.PG.Password),
		net.JoinHostPort(cfg.PG.Host, cfg.PG.Port),
		cfg.PG.DB,
	)

	return cfg, nil
}

func requireEnv(storage *string, envName string) error {
	if *storage != "" {
		return nil
	}

	val := os.Getenv(envName)
	if val == "" {
		return fmt.Errorf("environment variable %s not set", envName)
	}
	*storage = val
	return nil
}
