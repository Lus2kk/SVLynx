package apperrors

import "errors"

var (
	ErrInternalError   = errors.New("internal error")
	ErrSessionNotFound = errors.New("the session was not found or expired")
	ErrCodeExpired     = errors.New("the code has expired, request a new one")
	ErrInvalidCode     = errors.New("invalid code")
	ErrEmailCooldown   = errors.New("wait 60 seconds before resending the code")
	ErrCodeCooldown	   = errors.New("wait 5 seconds before resending the code")
	ErrTooManyAttempts = errors.New("exceeded the number of attempts. Try again later")
	ErrEmailSendFailed = errors.New("error when sending the code to the mail")
	ErrSessionCreate   = errors.New("failed to create a session")
	ErrUnauthorized    = errors.New("the user is not logged in")

	ErrNicknameExists = errors.New("the nickname is already occupied")
	ErrUserNotFound   = errors.New("the user was not found")
	ErrSaveUserFailed = errors.New("couldn't save profile")
)
