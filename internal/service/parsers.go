package service

import (
	"fmt"

	code "github.com/JanMird/auth/pkg/error"
	"github.com/JanMird/auth/internal/config"
	"github.com/JanMird/auth/internal/domain"
	"github.com/golang-jwt/jwt"
)

func parseRefreshToken(tokenString string) (domain.RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domain.RefreshClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.TokenSignString), nil
		},
	)

	if err != nil {
		return domain.RefreshClaims{}, fmt.Errorf("failed to parse token: %w", err)
	}
	if !token.Valid {
		return domain.RefreshClaims{}, fmt.Errorf("token is invalid")
	}

	claimsPtr, ok := token.Claims.(*domain.RefreshClaims)
	if !ok {
		return domain.RefreshClaims{}, fmt.Errorf("invalid token claims")
	}

	return *claimsPtr, nil
}

func parseAccessToken(tokenString string) (domain.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domain.Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.TokenSignString), nil
		},
	)

	if err != nil {
		return domain.Claims{}, fmt.Errorf("failed to parse token: %w", err)
	}
	if !token.Valid {
		return domain.Claims{}, fmt.Errorf("token is invalid: %w", code.ErrExpired)
	}

	claimsPtr, ok := token.Claims.(*domain.Claims)
	if !ok {
		return domain.Claims{}, fmt.Errorf("invalid token claims")
	}

	return *claimsPtr, nil
}
