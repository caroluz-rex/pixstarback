// middlewares/jwt_middleware.go

package middlewares

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt"
)

type contextKey string

const (
	ContextKeyPublicKey = contextKey("publicKey")
)

var jwtSecret = []byte("your_jwt_secret_key") // Должен совпадать с секретом из authentication.go

func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenStr := cookie.Value

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// Проверяем метод подписи
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrAbortHandler
			}
			return jwtSecret, nil
		})

		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			publicKey, ok := claims["publicKey"].(string)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Добавляем publicKey в контекст запроса
			ctx := context.WithValue(r.Context(), ContextKeyPublicKey, publicKey)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	})
}
