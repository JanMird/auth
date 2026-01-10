package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/JanMird/auth/internal/domain"
	code "github.com/JanMird/auth/pkg/error"
	"golang.org/x/crypto/bcrypt"
)

type Storage interface {
	GetUserWithEmail(context.Context, string) (domain.User, error)
	GetUserWithPhone(context.Context, string) (domain.User, error)
	GetUserWithID(context.Context, int) (domain.User, error)
	AddUser(context.Context, domain.User) (int, error)
	CheckUserEmailExists(context.Context, string) (bool, error)
	CheckUserPhoneExists(context.Context, string) (bool, error)
	AddRefreshToken(context.Context, string) error
	CheckRefreshToken(context.Context, string) (bool, error)
	RemoveRefreshToken(context.Context, string) error
}

type AuthService struct {
	Storage
}

func New(storage Storage) *AuthService {
	return &AuthService{
		Storage: storage,
	}
}

func (s *AuthService) Authenticate(ctx context.Context, user domain.UserUnsafe) (string, string, error) {
	var userDB domain.User
	var err error

	switch {
	case user.Email != "":
		userDB, err = s.Storage.GetUserWithEmail(ctx, user.Email)
	case user.Phone != "":
		userDB, err = s.Storage.GetUserWithPhone(ctx, user.Phone)
	default:
		return "", "", errors.New("failed to authenticate: email or phone must not be empty")
	}

	if err != nil {
		return "", "", err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(userDB.PasswordHashed), []byte(user.Password)); err != nil {
		return "", "", fmt.Errorf("invalid password: %w", err)
	}

	signedToken, err := createSignedAccessToken(userDB)
	if err != nil {
		return "", "", fmt.Errorf("failed to create signed access token: %w", err)
	}

	refreshSignedToken, err := createSignedRefreshToken(userDB.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to create signed refresh token: %w", err)
	}

	err = s.Storage.AddRefreshToken(ctx, refreshSignedToken)
	if err != nil {
		return "", "", err
	}

	return signedToken, refreshSignedToken, nil
}

func (s *AuthService) SignUp(ctx context.Context, user domain.UserUnsafe) (string, string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash password: %w", err)
	}

	userDB := domain.User{
		Email:          user.Email,
		Phone:          user.Phone,
		PasswordHashed: string(hashedPassword),
	}

	switch {
	case userDB.Email != "":
		ok, err := s.CheckUserEmailExists(ctx, userDB.Email)
		if err != code.ErrUserNotFound {
			return "", "", code.ErrInternal
		}
		if ok {
			return "", "", code.ErrUserAlreadyExists
		}
	case userDB.Phone != "":
		ok, err := s.CheckUserPhoneExists(ctx, userDB.Phone)
		if err != code.ErrUserNotFound {
			return "", "", code.ErrInternal
		}
		if ok {
			return "", "", code.ErrUserAlreadyExists
		}
	}

	userDB.ID, err = s.AddUser(ctx, userDB)
	if err != nil {
		return "", "", fmt.Errorf("failed to add user to db: %w", err)
	}

	signedToken, err := createSignedAccessToken(userDB)
	if err != nil {
		return "", "", errors.Join(code.ErrTokenNotCreated, err)
	}

	refreshSignedToken, err := createSignedRefreshToken(userDB.ID)
	if err != nil {
		return "", "", errors.Join(code.ErrTokenNotCreated, err)
	}

	err = s.Storage.AddRefreshToken(ctx, refreshSignedToken)
	if err != nil {
		return "", "", err
	}

	return signedToken, refreshSignedToken, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refresh_token string) (string, string, error) {
	ok, err := s.Storage.CheckRefreshToken(ctx, refresh_token)
	if err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}
	if !ok {
		return "", "", fmt.Errorf("refresh token is canceled: %w", code.ErrExpired)
	}

	err = s.Storage.RemoveRefreshToken(ctx, refresh_token)
	if err != nil {
		return "", "", err
	}

	refresh_claims, err := parseRefreshToken(refresh_token)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse token: %w", err)
	}

	userDB, err := s.Storage.GetUserWithID(ctx, refresh_claims.Sub)
	if err != nil {
		return "", "", err
	}

	signedToken, err := createSignedAccessToken(userDB)
	if err != nil {
		return "", "", fmt.Errorf("failed to create signed access token: %w", err)
	}

	refreshSignedToken, err := createSignedRefreshToken(userDB.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to create signed refresh token: %w", err)
	}

	err = s.Storage.AddRefreshToken(ctx, refreshSignedToken)
	if err != nil {
		return "", "", err
	}

	return signedToken, refreshSignedToken, nil
}

func (s *AuthService) GetClaimsFromToken(ctx context.Context, access_token string) (domain.Claims, error) {
	claims, err := parseAccessToken(access_token)
	if err != nil {
		return domain.Claims{}, fmt.Errorf("failed to parse token: %w", err)
	}

	return claims, nil
}

func (s *AuthService) SignOut(ctx context.Context, token string) error {
	err := s.Storage.RemoveRefreshToken(ctx, token)
	if err != nil {
		return err
	}

	return nil
}
