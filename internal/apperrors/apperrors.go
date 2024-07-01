// Модуль apperrors содержит ошибки приложения

package apperrors

import (
	"fmt"
)

// OriginalURLAlreadyExists структура ошибки
type OriginalURLAlreadyExists struct {
	URL string
}

// Error возвращает ошибку, если пользователя не существует
func (e *OriginalURLAlreadyExists) Error() string {
	return fmt.Sprintf("original URL already exists: %s", e.URL)
}
