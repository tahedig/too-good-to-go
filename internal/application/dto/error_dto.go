package dto

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code      string `json:"code" example:"OTP_INVALID"`
	Message   string `json:"message" example:"the verification code is incorrect"`
	Details   any    `json:"details,omitempty" swaggertype:"object"`
	RequestID string `json:"request_id,omitempty" example:"3f2a9c0b7d1e4f6a8b2c4d6e8f0a1b3c"`
}
