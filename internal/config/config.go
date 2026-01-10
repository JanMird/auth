package config

import "time"

var (
	TokenDurationMinutes = 15 * time.Minute
	RefreshTokenDurationMinutes =  24 * time.Hour
)

const (
	TokenSignString = "aboba"
)