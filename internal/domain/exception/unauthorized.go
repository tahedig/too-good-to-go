package exception

import "net/http"

const CodeUnauthorized = "UNAUTHORIZED"

func NewUnauthorized(message string) AppError {
	return &appError{status: http.StatusUnauthorized, code: CodeUnauthorized, message: message}
}
