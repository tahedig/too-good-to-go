package exception

import "net/http"

const CodeConflict = "CONFLICT"

func NewConflict(message string) AppError {
	return &appError{status: http.StatusConflict, code: CodeConflict, message: message}
}
