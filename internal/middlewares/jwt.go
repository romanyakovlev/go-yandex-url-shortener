package middlewares

import (
	"context"
	"errors"
	"net/http"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/apperrors"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/jwt"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
	userService "github.com/romanyakovlev/go-yandex-url-shortener/internal/service/user"
)

type contextKey string

const userContextKey contextKey = "currentUser"

func SetUserInContext(ctx context.Context, user models.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func GetUserFromContext(ctx context.Context) (models.User, bool) {
	user, ok := ctx.Value(userContextKey).(models.User)
	return user, ok
}

func addTokenToResponseWriter(userService *userService.UserService, w http.ResponseWriter) (models.User, error) {
	user, err := userService.CreateUserWithToken()
	if err != nil {
		return models.User{}, err
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   user.Token,
		Expires: jwt.GetExpiresAt(user.Token).Time,
	})
	return user, nil
}

func processExistingToken(userService *userService.UserService, tokenValue string, w http.ResponseWriter) (models.User, error) {
	userID := jwt.GetUserID(tokenValue)
	if userID == -1 {
		return addTokenToResponseWriter(userService, w)
	}
	user, userExists := userService.GetUserByID(userID)
	if !userExists {
		return models.User{}, &apperrors.AuthError{
			Message:    "UserID is not found",
			StatusCode: http.StatusUnauthorized,
		}
	}
	return user, nil
}

func JWTMiddleware(userService *userService.UserService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var user models.User
			var err error

			token, tokenErr := r.Cookie("token")
			if r.URL.Path == "/api/user/urls" {
				if tokenErr != nil || token == nil {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				user, err = processExistingToken(userService, token.Value, w)
			} else if token != nil {
				user, err = processExistingToken(userService, token.Value, w)
			} else {
				user, err = addTokenToResponseWriter(userService, w)
			}

			if err != nil {
				var appErr *apperrors.AuthError
				if errors.As(err, &appErr) {
					http.Error(w, appErr.Error(), appErr.StatusCode)
				} else {
					http.Error(w, "An unexpected error occurred", http.StatusInternalServerError)
				}
				return
			}

			ctxWithUser := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctxWithUser))
		})
	}
}
