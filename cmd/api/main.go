package main

import (
	"github.com/Dimetrius-eng/learning-go-shop/internal/config"
	"github.com/Dimetrius-eng/learning-go-shop/internal/database"
	"github.com/Dimetrius-eng/learning-go-shop/internal/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	log := logger.New()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	mainDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get database connection")
	}

	defer mainDB.Close()
	gin.SetMode(cfg.Server.GinMode)

	log.Info().Msg("starting server")
}
