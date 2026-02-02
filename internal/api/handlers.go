package api

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"nex-server/internal/auth"
	"nex-server/internal/config"
	"nex-server/internal/models"
	"nex-server/internal/ws"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func SetupRoutes(r *gin.Engine, wsManager *ws.Manager) {
	r.POST("/v1/login", func(c *gin.Context) {
		var login models.LoginRequest
		if err := c.BindJSON(&login); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		if login.Username != config.Current.User.Username || login.Password != config.Current.User.Password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		token, err := auth.GenerateLoginToken(login.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token gen failed"})
			return
		}

		c.JSON(http.StatusOK, models.LoginResponse{Token: token})
	})

	r.GET("/v1/img/tmp/:encodedPath", func(c *gin.Context) {
		encodedPath := c.Param("encodedPath")
		decodedBytes, err := base64.URLEncoding.DecodeString(encodedPath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
			return
		}
		c.File(string(decodedBytes))
	})

	r.GET("/v1/websocket", func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth header"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth header"})
			return
		}

		claims, err := auth.ValidateToken(parts[1])
		if err != nil || claims.Type != "login" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		wsToken, err := auth.GenerateWSToken(claims.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ws token gen failed"})
			return
		}

		wsUUID := uuid.New().String()
		socketURL := fmt.Sprintf("wss://%s:%d/v1/monitor/%s/ws", config.Current.API.Host, config.Current.API.Port, wsUUID)

		c.JSON(http.StatusOK, models.WebSocketResponse{
			Object: "websocket_token",
			Data: struct {
				Token  string "json:\"token\""
				Socket string "json:\"socket\""
			}{
				Token:  wsToken,
				Socket: socketURL,
			},
		})
	})

	r.GET("/v1/monitor/:uuid/ws", func(c *gin.Context) {
		ws.ServeWS(wsManager, c)
	})
}
