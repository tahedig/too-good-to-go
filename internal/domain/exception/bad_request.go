package exception

import "net/http"

const CodeBadRequest = "BAD_REQUEST"

// NewBadRequest is for invalid input. details may be nil or e.g. []FieldError.
func NewBadRequest(message string, details any) AppError {
	return &appError{status: http.StatusBadRequest, code: CodeBadRequest, message: message, details: details}
}
