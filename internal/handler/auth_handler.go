package handler

import (
	"context"
	"errors"

	errcode "github.com/JanMird/auth/error"
	pb "github.com/JanMird/auth/gen/auth"
	"github.com/JanMird/auth/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	Authenticate(context.Context, *domain.UserUnsafe) (string, error)
	SignUp(context.Context, *domain.UserUnsafe) (string, error)
}

type Logger interface {
	Debug(...interface{})
	Error(...interface{})
	Info(...interface{})
}

type Handler struct {
	pb.UnimplementedAuthServiceServer
	AuthService
	Logger
}

func New(service AuthService, logger Logger) *Handler {
	return &Handler{
		AuthService: service,
		Logger:      logger,
	}
}

func (h *Handler) Authenticate(ctx context.Context, req *pb.AuthenticateRequest) (*pb.AuthenticateRespone, error) {
	if ok := validateAuthInput(req); !ok {
		return nil, status.Error(codes.InvalidArgument, "phone or email must be not empty")
	}

	input := domain.UserUnsafe{
		Password: req.Password,
	}

	if req.Email != nil {
		input.Email = *req.Email
	}
	if req.Phone != nil {
		input.Phone = *req.Phone
	}

	token, err := h.AuthService.Authenticate(ctx, &input)

	switch {
	case errors.Is(err, errcode.ErrUserNotFound):
		return nil, status.Error(codes.NotFound, err.Error())
	case err != nil:
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AuthenticateRespone{
		JwtToken: token,
	}, nil
}

func (h *Handler) SignUp(ctx context.Context, req *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	if ok := validateSignUpInput(req); !ok {
		return nil, status.Error(codes.InvalidArgument, "phone or email must be not empty, password must not be empty")
	}

	input := domain.UserUnsafe{
		Password: req.Password,
	}

	if req.Email != nil {
		input.Email = *req.Email
	}
	if req.Phone != nil {
		input.Phone = *req.Phone
	}

	token, err := h.AuthService.SignUp(ctx, &input)

	switch {
	case err != nil:
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SignUpResponse{
		JwtToken: token,
	}, nil
}

func validateAuthInput(req *pb.AuthenticateRequest) bool {
	switch {
	case req.Email == nil && req.Phone == nil:
		return false
	case req.Email != nil && *req.Email == "":
		return false
	case req.Phone != nil && *req.Phone == "":
		return false
	}

	return true
}

func validateSignUpInput(req *pb.SignUpRequest) bool {
	switch {
	case req.Email == nil && req.Phone == nil:
		return false
	case req.Email != nil && *req.Email == "":
		return false
	case req.Phone != nil && *req.Phone == "":
		return false
	case req.Password == "":
		return false
	}

	return true
}
