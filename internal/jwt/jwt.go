// Package jwt выполняет работу с JWT-токенами для авторизации/аутентификации.
package jwt

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// TokenExp Время жизни токена
const TokenExp = time.Hour * 24 * 7

// SecretKey Секретный ключ для подписи токена
const SecretKey = "supersecretkey"

// BytesSecretKey Байтовое представление секретного ключа
var BytesSecretKey = []byte(SecretKey)

// Кэш для хранения данных о токенах
var cachedMap = NewCache()

// Claims представляет собой структуру с данными, закодированными в JWT-токене.
type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID // Уникальный идентификатор пользователя
}

// Cache предоставляет механизм кэширования данных о токенах.
type Cache struct {
	sync.Map // Использование встроенной карты с синхронизацией для безопасного доступа из разных горутин
}

// NewCache создает новый экземпляр кэша.
func NewCache() *Cache {
	return &Cache{}
}

// Get извлекает данные из кэша по строке токена.
func (c *Cache) Get(tokenString string) (*cachedData, bool) {
	data, found := c.Load(tokenString)
	if !found {
		return nil, false
	}
	return data.(*cachedData), true
}

// Set сохраняет данные в кэш по строке токена.
func (c *Cache) Set(tokenString string, data *cachedData) {
	c.Store(tokenString, data)
}

// cachedData содержит закэшированные данные о токене.
type cachedData struct {
	claims *Claims    // Данные, закодированные в токене
	token  *jwt.Token // Сам JWT-токен
}

// BuildJWTString создает строку JWT-токена для указанного идентификатора пользователя.
func BuildJWTString(UserID uuid.UUID) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)), // Установка времени истечения токена
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

// getClaimsWithToken извлекает данные из токена.
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

// getCachedData извлекает закэшированные данные о токене.
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

// GetExpiresAt возвращает срок истечения токена, если он валиден.
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

// GetUserID возвращает uuid пользователя из токена, если он валиден.
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
