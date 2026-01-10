package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/JanMird/auth/internal/domain"
	codes "github.com/JanMird/auth/pkg/error"
)

type Storage struct {
	*sql.DB
}

func New(conn *sql.DB) *Storage {
	return &Storage{DB: conn}
}

func (s *Storage) GetUserWithEmail(ctx context.Context, email string) (domain.User, error) {
	var user domain.User

	getUserByEmailQuery := `
		SELECT 
    		id,
			role,
			password_hash
		FROM users
		WHERE email = $1;
	`
	err := s.DB.QueryRowContext(ctx, getUserByEmailQuery, email).Scan(&user.ID, &user.Role, &user.PasswordHashed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, errors.Join(err, codes.ErrUserNotFound)
		}

		return domain.User{}, errors.Join(codes.ErrInternal, err)
	}

	return user, nil
}

func (s *Storage) GetUserWithPhone(ctx context.Context, phone string) (domain.User, error) {
	var user domain.User

	getUserByPhoneQuery := `
		SELECT 
    		id,
			role,
			password_hash
		FROM users
		WHERE phone = $1;
	`
	err := s.DB.QueryRowContext(ctx, getUserByPhoneQuery, phone).Scan(&user.ID, &user.Role, &user.PasswordHashed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, errors.Join(err, codes.ErrUserNotFound)
		}

		return domain.User{}, errors.Join(codes.ErrInternal, err)
	}

	return user, nil
}

func (s *Storage) GetUserWithID(ctx context.Context, id int) (domain.User, error) {
	var user domain.User

	getUserByPhoneQuery := `
		SELECT 
    		id,
			role,
			password_hash
		FROM users
		WHERE id = $1;
	`
	err := s.DB.QueryRowContext(ctx, getUserByPhoneQuery, id).Scan(&user.ID, &user.Role, &user.PasswordHashed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, errors.Join(err, codes.ErrUserNotFound)
		}

		return domain.User{}, errors.Join(codes.ErrInternal, err)
	}

	return user, nil
}

func (s *Storage) AddUser(ctx context.Context, u domain.User) (int, error) {
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

func (s *Storage) CheckUserEmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool

	checkUserExistsQuery := `
        SELECT EXISTS(
            SELECT 1 FROM users WHERE email = $1
        );
    `

	err := s.DB.QueryRowContext(ctx, checkUserExistsQuery, email).Scan(&exists)
	if err != nil {
		return false, errors.Join(codes.ErrInternal, err)
	}

	return exists, nil
}

func (s *Storage) CheckUserPhoneExists(ctx context.Context, phone string) (bool, error) {
	var exists bool

	checkUserExistsQuery := `
        SELECT EXISTS(
            SELECT 1 FROM users WHERE phone = $1
        );
    `

	err := s.DB.QueryRowContext(ctx, checkUserExistsQuery, phone).Scan(&exists)
	if err != nil {
		return false, errors.Join(codes.ErrInternal, err)
	}

	return exists, nil
}

func (s *Storage) AddRefreshToken(ctx context.Context, token string) error {
	addTokenQuery := `
        INSERT INTO ref_token (ref_token)
        VALUES ($1)
	`

	_, err := s.DB.ExecContext(ctx, addTokenQuery, token)
	if err != nil {
		return errors.Join(codes.ErrInternal, err)
	}

	return nil
}

func (s *Storage) CheckRefreshToken(ctx context.Context, token string) (bool, error) {
	var exists bool

	checkTokenExistsQuery := `
        SELECT EXISTS(
            SELECT 1 FROM ref_token WHERE ref_token = $1
        );
    `

	err := s.DB.QueryRowContext(ctx, checkTokenExistsQuery, token).Scan(&exists)
	if err != nil {
		return false, errors.Join(codes.ErrInternal, err)
	}

	return exists, nil
}

func (s *Storage) RemoveRefreshToken(ctx context.Context, token string) error {
	removeTokenQuery := `
        DELETE FROM ref_token 
        WHERE ref_token = $1
    `

	_, err := s.DB.ExecContext(ctx, removeTokenQuery, token)
	if err != nil {
		return errors.Join(err, codes.ErrInternal)
	}

	return nil
}
