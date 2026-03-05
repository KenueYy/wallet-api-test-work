package main

import (
	"log"
	"strconv"

	"github.com/KenueYy/wallet-api/internal/config"
	"github.com/KenueYy/wallet-api/internal/db"
	"github.com/KenueYy/wallet-api/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	if err := db.Init(cfg); err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	handlers.RegisterRoutes(r)

	log.Printf("Server port :%d", cfg.Port)
	if err := r.Run(":" + strconv.Itoa(cfg.Port)); err != nil {
		log.Fatal(err)
	}
}
