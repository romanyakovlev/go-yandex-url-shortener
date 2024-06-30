// Package jwt выполняет работу с JWT-токенами для авторизации/аутентификации
package jwt

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const (
	TokenExp  = time.Hour * 24 * 7
	SecretKey = "supersecretkey"
)

var (
	BytesSecretKey = []byte(SecretKey)
	cachedMap      = NewCache()
)

type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID
}

type Cache struct {
	sync.Map
}

func NewCache() *Cache {
	return &Cache{}
}

func (c *Cache) Get(tokenString string) (*cachedData, bool) {
	data, found := c.Load(tokenString)
	if !found {
		return nil, false
	}
	return data.(*cachedData), true
}

func (c *Cache) Set(tokenString string, data *cachedData) {
	c.Store(tokenString, data)
}

type cachedData struct {
	claims *Claims
	token  *jwt.Token
}

func BuildJWTString(UserID uuid.UUID) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: UserID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(BytesSecretKey)
	if err != nil {
		return "", err
	}

	cachedMap.Set(tokenString, &cachedData{claims: &claims, token: token})

	return tokenString, nil
}

func getClaimsWithToken(tokenString string) (*Claims, *jwt.Token, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return BytesSecretKey, nil
		})
	return claims, token, err
}

func getCachedData(tokenString string) (*cachedData, error) {
	cacheData, found := cachedMap.Get(tokenString)
	if found {
		if err := cacheData.token.Claims.Valid(); err != nil {
			return &cachedData{}, nil
		}
		return cacheData, nil
	}
	return &cachedData{}, nil
}

func GetExpiresAt(tokenString string) *jwt.NumericDate {
	data, err := getCachedData(tokenString)
	if err == nil {
		return data.claims.ExpiresAt
	}
	claims, token, err := getClaimsWithToken(tokenString)
	if err != nil || !token.Valid {
		return nil
	}

	cachedMap.Set(tokenString, &cachedData{claims: claims, token: token})

	return claims.ExpiresAt
}

func GetUserID(tokenString string) uuid.UUID {
	data, err := getCachedData(tokenString)
	if err == nil {
		return data.claims.UserID
	}

	claims, token, err := getClaimsWithToken(tokenString)
	if err != nil || !token.Valid {
		return uuid.UUID{}
	}

	cachedMap.Set(tokenString, &cachedData{claims: claims, token: token})

	return claims.UserID
}
