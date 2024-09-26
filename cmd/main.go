package main

import (
	"context"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/jamm3e3333/whalebone-go-test-project/cmd/app/config"
	"github.com/jamm3e3333/whalebone-go-test-project/cmd/app/setup/postgres"
	"github.com/jamm3e3333/whalebone-go-test-project/cmd/app/setup/prometheus"
	_ "github.com/jamm3e3333/whalebone-go-test-project/cmd/app/swagger"
	"github.com/jamm3e3333/whalebone-go-test-project/cmd/internal"
	"github.com/jamm3e3333/whalebone-go-test-project/cmd/internal/infrastructure/pg"
	"github.com/jamm3e3333/whalebone-go-test-project/cmd/internal/ui/http/health"
	healthcheck "github.com/jamm3e3333/whalebone-go-test-project/pkg/health"
	"github.com/jamm3e3333/whalebone-go-test-project/pkg/logger"
	pkgGin "github.com/jamm3e3333/whalebone-go-test-project/pkg/net/http/gin"
	"github.com/jamm3e3333/whalebone-go-test-project/pkg/net/http/ginprometheus"
	"github.com/jamm3e3333/whalebone-go-test-project/pkg/net/http/server"
	"github.com/jamm3e3333/whalebone-go-test-project/pkg/pgx"
	"github.com/jamm3e3333/whalebone-go-test-project/pkg/shutdown"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag"
)

// Swagger API setup:
// @title Whalebone Clients API
// @version 2.0
// @description API provides endpoints for whalebone clients
// @contact.name Whalebone
func main() {
	ctx := shutdown.SetupShutdownContext()

	var (
		appConfig, errAPPConfig       = config.CreateAPPConfig()
		loggerConfig, errLoggerConfig = config.CreateLoggerConfig()
		pgConfig, errPGConfig         = config.CreatePostgresConfig()
	)

	for _, err := range []error{
		errAPPConfig,
		errLoggerConfig,
		errPGConfig,
	} {
		if err != nil {
			panic(err)
		}
	}

	location, _ := time.LoadLocation(appConfig.Timezone)
	time.Local = location

	lg := logger.New(logger.ParseLevel(loggerConfig.Level), loggerConfig.DevMode)

	mm := prometheus.NewMetricsOnce(appConfig.AppName)()
	pc := postgres.EstablishConnection(ctx, pgx.Config{
		ConnectionURL:     pgConfig.ConnectionURL(),
		LogLevel:          pgConfig.LogLevel,
		MaxConnLifetime:   pgConfig.MaxConnLifetime,
		MaxConnIdleTime:   pgConfig.MaxConnIdleTIme,
		QueryTimeout:      pgConfig.QueryTimeout,
		DefaultMaxConns:   pgConfig.MaxConns,
		DefaultMinConns:   pgConfig.MinConns,
		HealthCheckPeriod: pgConfig.HealthCheckPeriod,
	}, lg, mm.Pm)

	// Http server
	lg.Info("Initializing http server...")

	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)

	ge := gin.New()
	ge.Use(gin.Recovery())
	pprof.Register(ge)

	// Register logger middleware
	ge.Use(
		cors.New(cors.Config{
			AllowOrigins:     appConfig.AllowedOrigins(),
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowCredentials: true,
		}),
		pkgGin.LoggerMiddleware(pkgGin.NewLoggerMiddlewareConfig(
			[]string{"/metrics", "/health/liveness", "/health/readiness", "/status", "/api/*any"},
		), lg),
	)

	// Register prometheus endpoint and request/response metrics
	ge.GET("/metrics", ginprometheus.Handler())
	ge.Use(ginprometheus.Measure(ginprometheus.Config{
		Subsystem: appConfig.AppName,
		Labels:    []ginprometheus.Label{},
	}))
	lg.Info("gin prometheus initialized")

	// Initialize Swagger
	gsc := ginSwagger.Config{
		URL:                      "doc.json",
		DocExpansion:             "list",
		InstanceName:             swag.Name,
		Title:                    "Whalebone Clients API",
		DefaultModelsExpandDepth: 2,
		DeepLinking:              true,
		PersistAuthorization:     false,
		Oauth2DefaultClientID:    "",
	}
	ge.GET("/api/*any", ginSwagger.CustomWrapHandler(&gsc, swaggerFiles.Handler))
	lg.Info("swagger initialized")

	// Initialize health check
	livenessHCh := healthcheck.NewHealthCheck(appConfig.HealthCheckTimeout, lg)
	livenessHCh.RegisterIndicator(pg.NewHealthIndicator(ctx, pc, lg))

	readinessHCh := healthcheck.NewHealthCheck(appConfig.HealthCheckTimeout, lg)

	hc := health.NewController(readinessHCh, livenessHCh, ge)
	hc.Register(ge)
	lg.Info("health check controller initialized")

	internal.RegisterModule(ge, internal.ModuleParams{
		PGConn: pc,
		Logger: lg,
		AppENV: appConfig.AppEnv,
	})

	for _, v := range ge.Routes() {
		lg.Info("[HTTP] Route: %s %s initialized.", v.Method, v.Path)
	}
	lg.Info("Internal module initialized.")
	lg.Info("[HTTP] Gin initialized.")

	srv := server.NewServer(ge, appConfig.ReadTimeout, appConfig.WriteTimeout, appConfig.Port, appConfig.ShutdownTimeout)
	lg.Info("[HTTP] Server initialized.")

	lg.Info("[HTTP] Start listening on port %d.", appConfig.Port)
	httpErrChan := srv.Run()

	select {
	case err := <-httpErrChan:
		lg.Error("http server error, %s", err)
		shutdown.SignalShutdown()
	case <-ctx.Done():
		if err := srv.Shutdown(context.Background()); err != nil {
			lg.Error("err shutting down http server, error: %v", err)
		}
		lg.Info("shutdown signaled")
	}
}
