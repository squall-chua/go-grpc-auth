package auth

import (
	"unicode"

	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ValidatePassword(password string, policy domain.PasswordPolicy) error {
	if len(password) < policy.MinLength {
		return status.Errorf(codes.InvalidArgument, "password must be at least %d characters", policy.MinLength)
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if policy.RequireUppercase && !hasUpper {
		return status.Error(codes.InvalidArgument, "password must contain at least one uppercase letter")
	}
	if policy.RequireLowercase && !hasLower {
		return status.Error(codes.InvalidArgument, "password must contain at least one lowercase letter")
	}
	if policy.RequireNumber && !hasNumber {
		return status.Error(codes.InvalidArgument, "password must contain at least one number")
	}
	if policy.RequireSpecial && !hasSpecial {
		return status.Error(codes.InvalidArgument, "password must contain at least one special character")
	}

	return nil
}
