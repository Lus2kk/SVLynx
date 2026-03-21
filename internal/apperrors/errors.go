package apperrors

import "errors"

var (
	ErrInvalidHash = errors.New("invalid telegram hash signature") 
	ErrAuthExpired     = errors.New("auth data expired, please try again")
    ErrSessionNotFound = errors.New("session not found or already used")
    ErrInvalidRequest  = errors.New("invalid request body")
    ErrInternal        = errors.New("internal server error")
)