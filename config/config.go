package config

import (
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
)

type (
	Config struct {
		REST
		PG
	}

	REST struct {
		Port string `setEnv:"PORT"`
		Host string `setEnv:"HOST"`
	}

	PG struct {
		URL      string
		Host     string `setEnv:"POSTGRES_HOST"`
		Port     string `setEnv:"POSTGRES_PORT"`
		DB       string `setEnv:"POSTGRES_DB"`
		User     string `setEnv:"POSTGRES_USER"`
		Password string `setEnv:"POSTGRES_PASSWORD"`
	}
)

func New() (*Config, error) {
	if err := godotenv.Load(".env"); err != nil {
		return nil, fmt.Errorf("failed to load .env: %w", err)
	}

	cfg := &Config{}

	pflag.StringVar(&cfg.REST.Port, "port", cfg.REST.Port, "port to listen on")
	pflag.Parse()

	envVars := map[string]*string{
		"REST_PORT":         &cfg.REST.Port,
		"REST_HOST":         &cfg.REST.Host,
		"POSTGRES_HOST":     &cfg.PG.Host,
		"POSTGRES_PORT":     &cfg.PG.Port,
		"POSTGRES_DB":       &cfg.PG.DB,
		"POSTGRES_USER":     &cfg.PG.User,
		"POSTGRES_PASSWORD": &cfg.PG.Password,
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
