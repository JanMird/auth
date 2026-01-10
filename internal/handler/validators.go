package handler

import pb "github.com/JanMird/auth/gen/auth"

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
