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

type AuthError struct {
	StatusCode int
	Message    string
}

func (e *AuthError) Error() string {
	return e.Message
}
