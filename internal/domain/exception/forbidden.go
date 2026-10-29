package exception

import "net/http"

const CodeForbidden = "FORBIDDEN"

func NewForbidden(message string) AppError {
	return &appError{status: http.StatusForbidden, code: CodeForbidden, message: message}
}
