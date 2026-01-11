package handler

import (
	"context"
	"errors"
	"strconv"

	"github.com/JanMird/auth/gen/auth"
	pb "github.com/JanMird/auth/gen/auth"
	"github.com/JanMird/auth/internal/domain"
	errcode "github.com/JanMird/auth/pkg/error"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthService interface {
	Authenticate(context.Context, domain.UserUnsafe) (string, string, error)
	SignUp(context.Context, domain.UserUnsafe) (string, string, error)
	RefreshToken(context.Context, string) (string, string, error)
	GetClaimsFromToken(context.Context, string) (domain.Claims, error)
	SignOut(context.Context, string) error
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

	token, refreshToken, err := h.AuthService.Authenticate(ctx, input)

	switch {
	case errors.Is(err, errcode.ErrUserNotFound):
		return nil, status.Error(codes.NotFound, err.Error())
	case errors.Is(err, errcode.ErrTokenNotCreated):
		return nil, status.Error(codes.DataLoss, err.Error())
	case err != nil:
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AuthenticateRespone{
		AccessToken:  token,
		RefreshToken: refreshToken,
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

	token, refreshToken, err := h.AuthService.SignUp(ctx, input)

	if err != nil {
		h.Logger.Error(err)
		switch {
		case errors.Is(err, errcode.ErrTokenNotCreated):
			h.Logger.Error("token not created")
			return nil, status.Error(codes.DataLoss, err.Error())
		case err != nil:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.SignUpResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
	}, nil
}

func (h *Handler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	token, refreshToken, err := h.AuthService.RefreshToken(ctx, req.RefreshToken)

	switch {
	case errors.Is(err, errcode.ErrTokenNotCreated):
		return nil, status.Error(codes.DataLoss, err.Error())
	case err != nil:
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RefreshTokenResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
	}, nil
}

func (h *Handler) GetClaimsFromToken(ctx context.Context, req *pb.GetClaimsFromTokenRequest) (*pb.GetClaimsFromTokenResponse, error) {
	claims, err := h.AuthService.GetClaimsFromToken(ctx, req.AccessToken)

	switch {
	case errors.Is(err, errcode.ErrExpired):
		return nil, status.Error(codes.Unauthenticated, err.Error())
	case err != nil:
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetClaimsFromTokenResponse{
		Id:   strconv.Itoa(claims.Sub),
		Role: auth.Role(claims.Role),
	}, nil
}

func (h *Handler) SignOut(ctx context.Context, req *pb.SignOutRequest) (*emptypb.Empty, error) {
	err := h.AuthService.SignOut(ctx, req.Token)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
