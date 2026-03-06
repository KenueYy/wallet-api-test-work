package main

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/KenueYy/wallet-api/internal/config"
	"github.com/KenueYy/wallet-api/internal/db"
	"github.com/KenueYy/wallet-api/internal/handlers"

	"github.com/gin-gonic/gin"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {
	cfg := config.Load()

	if err := db.Init(cfg); err != nil {
		logger.Error("failed to initialize database",
			"error", err.Error(),
		)
		os.Exit(1)
	}

	r := gin.Default()
	handlers.RegisterRoutes(r)

	logger.Info("server starting",
		"port", cfg.Port,
	)

	if err := r.Run(":" + strconv.Itoa(cfg.Port)); err != nil {
		logger.Error("server stopped with error",
			"error", err.Error(),
		)
		os.Exit(1)
	}
}
