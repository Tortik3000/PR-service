package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tortik3000/PR-service/db"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/Tortik3000/PR-service/config"
	api "github.com/Tortik3000/PR-service/generated/api/pr-service"
	controller "github.com/Tortik3000/PR-service/internal/controller/pr-service"
	repositury "github.com/Tortik3000/PR-service/internal/repository/pr-service"
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

	//shutdown := initTracer(logger, cfg.Observability.JaegerURL)
	//defer func() {
	//	err := shutdown(ctx)
	//
	//	if err != nil {
	//		logger.Error("can not shutdown jaeger collector", zap.Error(err))
	//	}
	//}()
	//
	//go runMetricsServer(logger, cfg.Observability.MetricsPort)
	//runPyroscope(logger, cfg.Observability.PyroscopeUrl)

	dbPool := initDBPool(cfg, logger)
	defer dbPool.Close()

	db.SetupPostgres(dbPool, logger)

	//db.SetupPostgres(dbPool, logger)

	repo := repositury.NewPostgresRepo(logger, dbPool)
	//transactor := repository.NewTransactor(dbPool, logger)
	useCases := usecase.NewUseCase(logger, repo, repo, repo)
	ctrl := controller.NewPRService(logger, useCases, useCases, useCases)
	r := chi.NewMux()

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

	r.Use(OpenAPIValidatorMiddleware(router))

	serverInterface := api.NewStrictHandler(ctrl, nil)
	h := api.HandlerFromMux(serverInterface, r)

	srv := &http.Server{
		Handler: h,
		Addr:    net.JoinHostPort(cfg.REST.Host, cfg.REST.Port),
	}

	go gracefulShutdown(ctx, srv, logger)

	logger.Info("Server started", zap.String("address", srv.Addr))
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal("Server start error", zap.Error(err))
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

func OpenAPIValidatorMiddleware(router routers.Router) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route, pathParams, err := router.FindRoute(r)
			if err != nil {
				http.Error(w, "Route not found", http.StatusNotFound)
				return
			}

			requestValidationInput := &openapi3filter.RequestValidationInput{
				Request:    r,
				PathParams: pathParams,
				Route:      route,
			}

			if err = openapi3filter.ValidateRequest(r.Context(), requestValidationInput); err != nil {
				http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
