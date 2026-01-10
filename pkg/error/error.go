package error

import "errors"

var (
	ErrUserNotFound      error = errors.New("user not found")
	ErrCreateUserFailure error = errors.New("cannot create user in database")
	ErrInvalidInput      error = errors.New("input data is ivalid")
	ErrCreationTime      error = errors.New("creation time cannot be in the future")
	ErrExpired           error = errors.New("is expired")
	ErrInternal          error = errors.New("internal")
	ErrUserAlreadyExists error = errors.New("user already exists")
	ErrTokenNotCreated   error = errors.New("token not created")
)
