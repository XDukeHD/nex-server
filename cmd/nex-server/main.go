package main

import (
	"fmt"
	"log"
	"nex-server/internal/api"
	"nex-server/internal/config"
	"nex-server/internal/ws"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if !config.Current.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	wsManager := ws.NewManager()

	go wsManager.Run()

	api.SetupRoutes(r, wsManager)

	addr := fmt.Sprintf("%s:%d", config.Current.API.Host, config.Current.API.Port)
	log.Printf("Starting Server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
