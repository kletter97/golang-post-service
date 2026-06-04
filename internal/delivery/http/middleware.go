package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5" // или библиотека, которую ты используешь
)

type contextKey string

const UserIDKey contextKey = "userID"

// AuthMiddleware принимает секретный ключ для валидации JWT
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// достаем заголовок Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing authorization token", http.StatusUnauthorized)
				return
			}

			// токен передается в формате "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
				return
			}
			tokenString := parts[1]

			// валидируем токен
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// достаем id пользователя из Claims (полезной нагрузки токена)
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			var userID int64
			if idFloat, ok := claims["user_id"].(float64); ok {
				userID = int64(idFloat)
			} else {
				http.Error(w, "User ID missing in token", http.StatusUnauthorized)
				return
			}

			// кладем UserID в контекст запроса и передаем хэндлеру дальше по цепочке
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
