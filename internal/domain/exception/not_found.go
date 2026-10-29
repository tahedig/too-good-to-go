package exception

import "net/http"

const CodeNotFound = "NOT_FOUND"

func NewNotFound(message string) AppError {
	return &appError{status: http.StatusNotFound, code: CodeNotFound, message: message}
}
