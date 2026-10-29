package exception

import "net/http"

const CodeInternalServer = "INTERNAL_SERVER_ERROR"

func NewInternalServer(cause error) AppError {
	return &appError{
		status:  http.StatusInternalServerError,
		code:    CodeInternalServer,
		message: "an unexpected error occurred",
		cause:   cause,
	}
}
