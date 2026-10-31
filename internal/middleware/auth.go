package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"uuid"

	"ya_url_shortener/internal/config"
)

func signUserID(userID uuid.UUID, secretKey config.AuthSecretKey) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(userID.String()))
	signature := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("%s.%s", userID.String(), signature)
}

func parseToken(userToken string) (uuid.UUID, []byte, error) {
	parts := strings.Split(userToken, ".")
	if len(parts) != 2 { //nolint:mnd
		return uuid.UUID{}, nil, errors.New("invalid token size")
	}
	userID, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.UUID{}, nil, errors.New("invalid user id")
	}
	signature, err := hex.DecodeString(parts[1])
	if err != nil {
		return uuid.UUID{}, nil, errors.New("invalid signature")
	}
	return userID, signature, nil
}

func signatureValid(userID uuid.UUID, signature []byte, secretKey config.AuthSecretKey) bool {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(userID.String()))
	return hmac.Equal(signature, h.Sum(nil))
}

func newAuthCookie(userID uuid.UUID, secretKey config.AuthSecretKey) *http.Cookie {
	return &http.Cookie{ //nolint:gosec
		Name:     string(config.UserIDKey),
		Value:    signUserID(userID, secretKey),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
}

// resolveUserID возвращает userID для запроса.
// Нет куки или неверная подпись → новый пользователь (ok=true).
func resolveUserID(
	r *http.Request,
	userIDKey config.UserIdentifier,
	secretKey config.AuthSecretKey,
	logger *slog.Logger,
) (uuid.UUID, bool) {
	userCookie, err := r.Cookie(string(userIDKey))
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return uuid.New(), true
		}
		logger.ErrorContext(r.Context(), "Auth Error", "err", err)
		return uuid.UUID{}, false
	}

	userID, signature, err := parseToken(userCookie.Value)
	if err != nil {
		logger.ErrorContext(r.Context(), "Auth Error", "err", err)
		return uuid.UUID{}, false
	}
	if !signatureValid(userID, signature, secretKey) {
		return uuid.New(), true
	}
	return userID, true
}

// CookieAuth - middleware для аутентификации по куки userId, в котором должен быть UUID пользователя.
// Если куки нет, пользователь пришел впервые, создаем идентификатор и подставляем в куки.
// Если куки есть, но в нем нет идендитификатора пользователя, то отдаем 401.
func CookieAuth(
	userIDKey config.UserIdentifier,
	secretKey config.AuthSecretKey,
	logger *slog.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := resolveUserID(r, userIDKey, secretKey, logger)
			if !ok {
				http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
				return
			}

			http.SetCookie(w, newAuthCookie(userID, secretKey))
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
