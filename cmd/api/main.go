package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Dimetrius-eng/learning-go-shop/internal/config"
	"github.com/Dimetrius-eng/learning-go-shop/internal/database"
	"github.com/Dimetrius-eng/learning-go-shop/internal/events"
	"github.com/Dimetrius-eng/learning-go-shop/internal/interfaces"
	"github.com/Dimetrius-eng/learning-go-shop/internal/logger"
	"github.com/Dimetrius-eng/learning-go-shop/internal/providers"
	"github.com/Dimetrius-eng/learning-go-shop/internal/repositories"
	"github.com/Dimetrius-eng/learning-go-shop/internal/server"
	"github.com/Dimetrius-eng/learning-go-shop/internal/services"
	"github.com/gin-gonic/gin"
)

// @title E-Commerce API
// @version 1.0
// @description A modern e-commerce API built with Go, Gin, and GORM
// @termsOfService http://swagger.io/terms/

// @contact.name   Dmytro Rudenko
// @contact.url    https://www.linkedin.com/in/dima-rudenko-86b9633b7
// @contact.email  mydimka08@gmail.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemas http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	log := logger.New()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	mainDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get database connection")
	}
	fmt.Printf("%T\n", mainDB)
	defer mainDB.Close() //nolint:errcheck

	ctx := context.Background()

	eventPublisher, err := events.NewEventPublisher(ctx, &cfg.AWS)
	if err != nil {
		log.Error().Err(err).Msg("failed to create event publisher")
		return
	}
	gin.SetMode(cfg.Server.GinMode)

	userRepo := repositories.NewUserRepository(db)
	cartRepo := repositories.NewCartRepository(db)
	authService := services.NewAuthService(
		cfg,
		eventPublisher,
		userRepo,
		cartRepo,
	)
	productService := services.NewProductService(db)
	userService := services.NewUserService(db)
	cartService := services.NewCartService(db)
	orderService := services.NewOrderService(db)

	var uploadProvider interfaces.UploadProvider
	if cfg.Upload.UploadProvider == "s3" {
		uploadProvider = providers.NewS3Provider(cfg)
	} else {
		uploadProvider = providers.NewLocalUploadProvider(cfg.Upload.Path)
	}
	uploadService := services.NewUploadService(uploadProvider)

	srv := server.New(cfg,
		&log,
		authService,
		productService,
		userService,
		uploadService,
		cartService,
		orderService)

	router := srv.SetupRoutes()

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Info().Str("port", cfg.Server.Port).Msg("starting http server")
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("failed to start http server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("failed to shutdown http server")
		return
	}

	log.Info().Msg("shutting down database")
}
