package domain

import (
	"errors"
	"time"

	code "github.com/JanMird/auth/error"
)

type Claims struct {
	Sub  int
	Role int
	Exp  time.Duration
	Iat  time.Time
}

func (c *Claims) Valid() error {
	if time.Now().Before(c.Iat) {
		return code.ErrCreationTime
	}

	if c.Iat.Add(c.Exp).Before(time.Now()) {
		return errors.Join(code.ErrExpired, errors.New("token is expired"))
	}

	return nil
}
