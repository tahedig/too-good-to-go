package exception

import "net/http"

const CodeMethodNotAllowed = "METHOD_NOT_ALLOWED"

func NewMethodNotAllowed(message string) AppError {
	return &appError{status: http.StatusMethodNotAllowed, code: CodeMethodNotAllowed, message: message}
}
