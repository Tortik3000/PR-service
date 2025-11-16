package app

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/Tortik3000/PR-service/config"
	"github.com/Tortik3000/PR-service/db"
	api "github.com/Tortik3000/PR-service/generated/api/pr-service"
	controller "github.com/Tortik3000/PR-service/internal/controller/pr-service"
	"github.com/Tortik3000/PR-service/internal/metrics"
	repoMiddlerware "github.com/Tortik3000/PR-service/internal/middleware/repo_middleware"
	restMiddlerware "github.com/Tortik3000/PR-service/internal/middleware/rest_middleware"
	repository "github.com/Tortik3000/PR-service/internal/repository/pr-service"
	usecase "github.com/Tortik3000/PR-service/internal/usecase/pr-service"
)

const (
	gracefulShutdownTimeout = 5 * time.Second
	writeTimeout            = 10 * time.Second
	readTimeout             = 10 * time.Second
	DBMaxConnections        = 10
	DBMinConnections        = 2
	DBMaxConnLifetime       = time.Hour
	DBMaxConnIdleTime       = 30 * time.Minute
	DBOperationTimeLimit    = 5 * time.Second
	DBMaxAttemptsForPing    = 5
	DBDelayForPing          = 2 * time.Second
)

func Run(
	logger *zap.Logger,
	cfg *config.Config,
) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dbPool := initDBPool(cfg, logger)
	defer dbPool.Close()

	db.SetupPostgres(dbPool, logger)

	repo := repository.NewPostgresRepo(logger, dbPool)
	metricsRepo := repoMiddlerware.NewMiddlewareMetricsRepo(repo, metrics.DBQueryLatency)

	transactor := repository.NewTransactor(dbPool, logger)
	useCases := usecase.NewUseCase(logger, metricsRepo, metricsRepo, metricsRepo, transactor)
	ctrl := controller.NewPRService(logger, useCases, useCases, useCases)

	go runMetricsServer(ctx, logger, cfg.Observability.MetricsPort)
	runPRServer(ctx, logger, ctrl, cfg)
}

func runPRServer(ctx context.Context, logger *zap.Logger, ctrl api.StrictServerInterface, cfg *config.Config) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("api/pr-service/pr-service.yml")
	if err != nil {
		logger.Fatal("Failed to load OpenAPI spec", zap.Error(err))
	}

	if err = doc.Validate(context.Background()); err != nil {
		logger.Fatal("OpenAPI spec validation failed", zap.Error(err))
	}

	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		return
	}

	r := chi.NewMux()

	r.Use(restMiddlerware.MetricsMiddleware("pr-service"))
	r.Use(restMiddlerware.OpenAPIValidatorMiddleware(router))

	serverInterface := api.NewStrictHandler(ctrl, nil)
	h := api.HandlerFromMux(serverInterface, r)

	srv := &http.Server{
		Handler: h,
		Addr:    ":" + cfg.REST.Port,
	}

	go gracefulShutdown(ctx, srv, logger)

	logger.Info("Server started", zap.String("address", srv.Addr))
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal("Server start error", zap.Error(err))
	}
}

func runMetricsServer(ctx context.Context, logger *zap.Logger, port string) {
	r := chi.NewRouter()
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go gracefulShutdown(ctx, srv, logger)

	logger.Info("starting metrics server", zap.String("port", port))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal("can not start metrics server", zap.Error(err))
	}
}

func gracefulShutdown(ctx context.Context, srv *http.Server, logger *zap.Logger) {
	<-ctx.Done()
	logger.Info("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server shutdown failed: ", zap.Error(err))
	}
	logger.Info("Shutting down service-profile")
}

func initDBPool(cfg *config.Config, logger *zap.Logger) *pgxpool.Pool {
	var dbPool *pgxpool.Pool

	pgxCfg, err := pgxpool.ParseConfig(cfg.PG.URL)
	if err != nil {
		logger.Fatal("Unable to parse connection string", zap.Error(err))
	}

	pgxCfg.MaxConns = DBMaxConnections
	pgxCfg.MinConns = DBMinConnections
	pgxCfg.MaxConnLifetime = DBMaxConnLifetime
	pgxCfg.MaxConnIdleTime = DBMaxConnIdleTime

	ctx, cancel := context.WithTimeout(context.Background(), DBOperationTimeLimit)
	defer cancel()

	dbPool, err = pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		logger.Fatal("Unable to create connection pool", zap.Error(err))
	}

	for attempt := range DBMaxAttemptsForPing {
		ctxPing, cancel := context.WithTimeout(context.Background(), DBOperationTimeLimit)
		err = dbPool.Ping(ctxPing)
		cancel()

		if err == nil {
			break
		}

		logger.Warn("Database ping failed", zap.Int("attempt", attempt), zap.Error(err))

		if attempt < DBMaxAttemptsForPing {
			time.Sleep(DBDelayForPing)
		}
	}

	if err != nil {
		logger.Fatal("Unable to ping database after several attempts", zap.Error(err))
	}

	logger.Info("Database connection pool established")
	return dbPool
}
