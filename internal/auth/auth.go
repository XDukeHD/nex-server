package auth

import (
	"errors"
	"nex-server/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Username string `json:"username"`
	Type     string `json:"type"`
	jwt.RegisteredClaims
}

func GenerateLoginToken(username string) (string, error) {
	expirationTime := time.Now().Add(30 * 24 * time.Hour)
	claims := &Claims{
		Username: username,
		Type:     "login",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Current.JWTSecret))
}

func GenerateWSToken(username string) (string, error) {
	expirationTime := time.Now().Add(20 * time.Minute)
	claims := &Claims{
		Username: username,
		Type:     "websocket",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Current.JWTSecret))
}

func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Current.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
