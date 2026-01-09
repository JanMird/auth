package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	codes "github.com/JanMird/auth/error"
	"github.com/JanMird/auth/internal/domain"
)

type Storage struct {
	*sql.DB
}

func New(conn *sql.DB) *Storage {
	return &Storage{DB: conn}
}

func (s *Storage) GetUserWithEmail(ctx context.Context, email string) (*domain.User, error) {
	var ans domain.User

	getUserByEmailQuery := `
		SELECT 
    		id,
			role,
			password_hash
		FROM users
		WHERE email = $1;
	`
	err := s.DB.QueryRowContext(ctx, getUserByEmailQuery, email).Scan(&ans.ID, &ans.Role, &ans.PasswordHashed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Join(err, codes.ErrUserNotFound)
		}

		return nil, err
	}

	return &ans, nil
}

func (s *Storage) GetUserWithPhone(ctx context.Context, phone string) (*domain.User, error) {
	var ans domain.User

	getUserByPhoneQuery := `
		SELECT 
    		id,
			role,
			password_hash
		FROM users
		WHERE phone = $1;
	`
	err := s.DB.QueryRowContext(ctx, getUserByPhoneQuery, phone).Scan(&ans.ID, &ans.Role, &ans.PasswordHashed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Join(err, codes.ErrUserNotFound)
		}

		return nil, err
	}

	return &ans, nil
}

func (s *Storage) GetUserWithID(ctx context.Context, id string) (*domain.User, error) {
	var ans domain.User

	getUserByPhoneQuery := `
		SELECT 
    		id,
			role,
			password_hash
		FROM users
		WHERE id = $1;
	`
	err := s.DB.QueryRowContext(ctx, getUserByPhoneQuery, id).Scan(&ans.ID, &ans.Role, &ans.PasswordHashed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Join(err, codes.ErrUserNotFound)
		}

		return nil, err
	}

	return &ans, nil
}

func (s *Storage) AddUser(ctx context.Context, u *domain.User) (int, error) {
	var id int
	now := time.Now()

	addUserQuery := `
        INSERT INTO users (email, password_hash, role, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id;
	`

	err := s.DB.QueryRowContext(ctx, addUserQuery, u.Email, u.PasswordHashed, u.Role, now, now).Scan(&id)
	if err != nil {
		return -1, codes.ErrCreateUserFailure
	}

	return id, nil
}
