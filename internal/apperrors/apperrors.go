// Модуль apperrors содержит ошибки приложения

package apperrors

import (
	"fmt"
)

type OriginalURLAlreadyExists struct {
	URL string
}

func (e *OriginalURLAlreadyExists) Error() string {
	return fmt.Sprintf("original URL already exists: %s", e.URL)
}
