package interceptors

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/romanyakovlev/go-yandex-url-shortener/internal/jwt"
	"github.com/romanyakovlev/go-yandex-url-shortener/internal/models"
)

type contextKey string

const userContextKey contextKey = "currentUser"

// GetUserFromContext получает пользователя из контекста запроса
func GetUserFromContext(ctx context.Context) (models.User, bool) {
	user, ok := ctx.Value(userContextKey).(models.User)
	return user, ok
}

func addTokenToMetadata(ctx context.Context) (models.User, error) {
	UUID := uuid.New()
	token, err := jwt.BuildJWTString(UUID)
	if err != nil {
		return models.User{}, err
	}
	user := models.User{
		UUID:  UUID,
		Token: token,
	}
	md := metadata.Pairs(
		"token", user.Token,
		"expires-at", jwt.GetExpiresAt(user.Token).Format(time.RFC3339),
	)
	err = grpc.SendHeader(ctx, md)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func processExistingToken(tokenValue string, ctx context.Context) (models.User, error) {
	userID := jwt.GetUserID(tokenValue)
	if userID == uuid.Nil {
		return addTokenToMetadata(ctx)
	}
	return models.User{UUID: userID}, nil
}

// JWTAuthInterceptor  обеспечивает аутентификацию пользователя
// с помощью JWT-токенов, хранящихся в метадате.
func JWTAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var user models.User
	var err error
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	tokenSlice, ok := md["token"]
	if info.FullMethod == "/shortener.URLShortenerService/GetURLByID" {
		if !ok || len(tokenSlice) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}
		user, err = processExistingToken(tokenSlice[0], ctx)
	} else if len(tokenSlice) > 0 {
		user, err = processExistingToken(tokenSlice[0], ctx)
	} else {
		user, err = addTokenToMetadata(ctx)
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "An unexpected error occurred")
	}

	ctxWithUser := context.WithValue(ctx, userContextKey, user)
	return handler(ctxWithUser, req)
}
