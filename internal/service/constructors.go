package service

import (
	"fmt"
	"time"

	"github.com/JanMird/auth/internal/config"
	"github.com/JanMird/auth/internal/domain"
	"github.com/golang-jwt/jwt"
)

func createToken(claims jwt.Claims) *jwt.Token {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
}

func createSignedAccessToken(user domain.User) (string, error) {
	claims := domain.Claims{
		Sub:  user.ID,
		Role: user.Role,
		Exp:  config.TokenDurationMinutes,
		Iat:  time.Now(),
	}

	token := createToken(&claims)
	signedToken, err := token.SignedString(config.TokenSignString)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func createSignedRefreshToken(id int) (string, error) {
	claims := domain.RefreshClaims{
		Sub: id,
		Exp: config.RefreshTokenDurationMinutes,
		Iat: time.Now(),
	}

	token := createToken(&claims)
	signedToken, err := token.SignedString(config.TokenSignString)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}
