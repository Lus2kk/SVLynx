package apperrors

import "errors"

var (
	ErrInternalError 	 = errors.New("внутренняя ошибка")
	ErrSessionNotFound   = errors.New("сессия не найдена или истекла")
	ErrCodeExpired       = errors.New("код истёк, запросите новый")
	ErrInvalidCode       = errors.New("неверный код")
	ErrCooldown          = errors.New("подождите 60 секунд перед повторной отправкой кода")
	ErrTooManyAttempts   = errors.New("превышено кол-во попыток. попробуйте позже")
	ErrEmailSendFailed   = errors.New("ошибка при отправке кода на почту")
	ErrSessionCreate     = errors.New("не удалось создать сессию")	
	ErrUnauthorized      = errors.New("пользователь не авторизован")

	ErrNicknameExists    = errors.New("никнейм уже занят")
	ErrUserNotFound      = errors.New("пользователь не найден")
	ErrSaveUserFailed    = errors.New("не удалось сохранить профиль")
)