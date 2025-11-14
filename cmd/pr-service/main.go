package main

import (
	"github.com/Tortik3000/PR-service/config"
	"github.com/Tortik3000/PR-service/internal/app"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
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
	//logger, err := NewFileLogger()
	//if err != nil {
	//	log.Fatalf("can not initialize logger: %s", err)
	//}

	app.Run(logger, cfg)
}

//func NewFileLogger() (*zap.Logger, error) {
//	const logFile = "/app/logs/library.log"
//
//	err := os.MkdirAll("/app/logs", 0755)
//	if err != nil {
//		return nil, err
//	}
//
//	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
//	if err != nil {
//		return nil, err
//	}
//
//	writeSyncer := zapcore.AddSync(file)
//	encoderCfg := zap.NewProductionEncoderConfig()
//	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
//	encoder := zapcore.NewJSONEncoder(encoderCfg)
//
//	core := zapcore.NewCore(encoder, writeSyncer, zap.InfoLevel)
//
//	return zap.New(core), nil
//}
