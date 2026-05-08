package auth_handler

import (
	"github.com/svlynx/messenger/internal/apperrors"
	"github.com/svlynx/messenger/internal/auth/auth_models"
)

func validateTelegramCallback(req *auth_models.TelegramCallbackRequest) error{
	if req.ID == 0 {
		return apperrors.ErrInvalidRequest
	}
	if req.FirstName == "" {
		return apperrors.ErrInvalidRequest
	}
	if req.AuthDate == 0 {
		return apperrors.ErrInvalidRequest
	}
	if req.Hash == "" {
		return apperrors.ErrInvalidRequest
	}
	return nil
} 