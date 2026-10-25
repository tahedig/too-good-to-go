package exception

import "net/http"

const CodeInternalServer = "INTERNAL_SERVER_ERROR"

// NewInternalServer wraps an unexpected error. The cause is logged by the
// error middleware but never sent to the client.
func NewInternalServer(cause error) AppError {
	return &appError{
		status:  http.StatusInternalServerError,
		code:    CodeInternalServer,
		message: "an unexpected error occurred",
		cause:   cause,
	}
}
