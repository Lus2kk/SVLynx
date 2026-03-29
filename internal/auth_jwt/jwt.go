package auth_jwt

import (
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/svlynx/messenger/internal/apperrors"
	"github.com/svlynx/messenger/internal/auth_repository"
)

type Claims struct {
	Email 		string `json:"email"`
	TokenType	string `json:"token_type"`
	jwt.RegisteredClaims 
}

func generateToken(email, tokenType, secret string, ttl time.Duration) (string, error){
	claim := Claims{
		Email: email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString([]byte(secret))
}

func GenerateAccessToken(email, secret string) (string, error) {
	return generateToken(email, "access", secret, auth_repository.AccesTokenTTL)
}

func GenerateRefreshToken(email, secret string) (string, error) {
	return generateToken(email, "refresh", secret, auth_repository.RefreshTokenTTL)
}

func Parse(tokenSTR, secret string) (*Claims, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenSTR, &claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			slog.Warn("invalid signature method", "method", t.Method.Alg())
			return nil, apperrors.ErrInvalidSignature
		}
		return []byte(secret), nil
	})

	if err != nil {
		slog.Warn("token parsing error", "err", err)
		return nil, err
	}

	if !token.Valid {
		slog.Warn("invalid token")
		return nil, apperrors.ErrInvalidToken
	}

	slog.Info("token parsed successfully", "email", claims.Email, "token_type", claims.TokenType)

	return &claims, nil
}