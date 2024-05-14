package middlewares

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/jwt"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
)

type contextKey string

const userContextKey contextKey = "currentUser"

func GetUserFromContext(ctx context.Context) (models.User, bool) {
	user, ok := ctx.Value(userContextKey).(models.User)
	return user, ok
}

func addTokenToResponseWriter(w http.ResponseWriter) (models.User, error) {
	UUID := uuid.New()
	token, err := jwt.BuildJWTString(UUID)
	if err != nil {
		return models.User{}, err
	}
	user := models.User{
		UUID:  UUID,
		Token: token,
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   user.Token,
		Expires: jwt.GetExpiresAt(user.Token).Time,
	})
	return user, nil
}

func processExistingToken(tokenValue string, w http.ResponseWriter) (models.User, error) {
	userID := jwt.GetUserID(tokenValue)
	if userID == uuid.Nil {
		return addTokenToResponseWriter(w)
	}
	return models.User{UUID: userID}, nil
}

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		var err error

		token, tokenErr := r.Cookie("token")
		if r.URL.Path == "/api/user/urls" {
			if tokenErr != nil || token == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			user, err = processExistingToken(token.Value, w)
		} else if token != nil {
			user, err = processExistingToken(token.Value, w)
		} else {
			user, err = addTokenToResponseWriter(w)
		}

		if err != nil {
			http.Error(w, "An unexpected error occurred", http.StatusInternalServerError)
			return
		}

		ctxWithUser := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctxWithUser))
	})
}
