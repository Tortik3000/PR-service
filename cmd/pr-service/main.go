package main

import (
	"os"

	"github.com/Tortik3000/PR-service/config"
	"github.com/Tortik3000/PR-service/internal/app"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	cfg, err := config.New()

	if err != nil {
		log.Fatalf("can not get application config: %s", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can not initialize logger: %s", err)
	}

	app.Run(logger, cfg)
}
