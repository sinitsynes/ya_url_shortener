package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"uuid"

	"ya_url_shortener/internal/config"
)

// CookieAuth - middleware для аутентификации по куки userId, в котором должен быть UUID пользователя.
// Если куки нет, пользователь пришел впервые, создаем идентификатор и подставляем в куки.
// Если куки есть, но в нем нет идендитификатора пользователя, то отдаем 401.
func CookieAuth(userIDKey config.UserIdentifier, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCookie, err := r.Cookie(string(userIDKey))
			if err != nil {
				if errors.Is(err, http.ErrNoCookie) {
					userID := uuid.New()
					r = r.WithContext(context.WithValue(r.Context(), userIDKey, userID))
					next.ServeHTTP(w, r)
					return
				}
				logger.ErrorContext(r.Context(), "Auth Error", "err", err)
				http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
				return
			}
			idValue := userCookie.Value
			userID, err := uuid.Parse(idValue)
			if err != nil {
				logger.ErrorContext(r.Context(), "Auth Error", "err", err)
				http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), userIDKey, userID))
			next.ServeHTTP(w, r)
		},
		)
	}
}
