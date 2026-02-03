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

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Max-Age", "43200")
		c.Writer.Header().Set("Access-Control-Allow-Private-Network", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	wsManager := ws.NewManager()

	go wsManager.Run()

	api.SetupRoutes(r, wsManager)

	addr := fmt.Sprintf("%s:%d", config.Current.API.Host, config.Current.API.Port)
	log.Printf("Starting Server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
