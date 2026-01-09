package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	code "github.com/JanMird/auth/error"
	"github.com/JanMird/auth/internal/config"
	"github.com/JanMird/auth/internal/domain"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Storage interface {
	GetUserWithEmail(context.Context, string) (*domain.User, error)
	GetUserWithPhone(context.Context, string) (*domain.User, error)
	GetUserWithID(context.Context, string) (*domain.User, error)
	AddUser(context.Context, *domain.User) (int, error)
}

type AuthService struct {
	Storage
}

func New(storage Storage) *AuthService {
	return &AuthService{
		Storage: storage,
	}
}

func createToken(claims domain.Claims) *jwt.Token {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
}

func createSignedToken(user *domain.User) (string, error) {
	claims := domain.Claims{
		Sub:  user.ID,
		Role: user.Role,
		Exp:  config.TokenDurationMinutes * time.Minute,
		Iat:  time.Now(),
	}

	if claims.Valid() != nil {
		return "", fmt.Errorf("generated invalid token: %w", code.ErrInternal)
	}

	token := createToken(claims)
	signedToken, err := token.SignedString(config.TokenSignString)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (s *AuthService) Authenticate(ctx context.Context, user *domain.UserUnsafe) (string, error) {
	var userDB *domain.User
	var err error

	switch {
	case user.Email != "":
		userDB, err = s.Storage.GetUserWithEmail(ctx, user.Email)
	case user.Phone != "":
		userDB, err = s.Storage.GetUserWithPhone(ctx, user.Phone)
	}

	if err != nil {
		return "", err
	}
	if userDB == nil {
		return "", fmt.Errorf("cannot authenticate: %w", code.ErrInternal)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(userDB.PasswordHashed), []byte(user.Password)); err != nil {
		return "", fmt.Errorf("invalid password: %w", err)
	}

	signedToken, err := createSignedToken(userDB)
	if err != nil {
		return "", fmt.Errorf("failed to create signed token: %w", err)
	}

	return signedToken, nil
}

func (s *AuthService) SignUp(ctx context.Context, user *domain.UserUnsafe) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	userDB := domain.User{
		Email:          user.Email,
		Phone:          user.Phone,
		PasswordHashed: string(hashedPassword),
	}

	_, err = s.GetUserWithEmail(ctx, userDB.Email)
	if err != code.ErrUserNotFound {
		return "", code.ErrUserAlreadyExists
	}
	_, err = s.GetUserWithPhone(ctx, userDB.Phone)
	if err != code.ErrUserNotFound {
		return "", code.ErrUserAlreadyExists
	}

	signedToken, err := createSignedToken(&userDB)
	if err != nil {
		return "", errors.New("failed to create user (token creation)")
	}

	userDB.ID, err = s.AddUser(ctx, &userDB)
	if err != nil {
		return "", fmt.Errorf("failed to add user to db: %w", err)
	}

	return signedToken, nil
}
