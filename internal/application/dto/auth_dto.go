package dto

type RequestOTPRequest struct {
	// Iranian local (09...), 98... or E.164 (+...) format.
	Phone string `json:"phone" example:"09123456789"`
}

type RequestOTPResponse struct {
	Message          string `json:"message" example:"verification code sent"`
	ExpiresInSeconds int    `json:"expires_in_seconds" example:"300"`
	// DebugCode is present only when OTP_EXPOSE_CODE=true (never in production).
	DebugCode string `json:"debug_code,omitempty" example:"123456"`
}

type VerifyOTPRequest struct {
	Phone string `json:"phone" example:"09123456789"`
	Code  string `json:"code" example:"123456"`
}

type UserResponse struct {
	ID    string `json:"id" example:"5f0c6d1e-8b7a-4c1e-9a3f-2d4b6c8e0a12"`
	Phone string `json:"phone" example:"+989123456789"`
	Role  string `json:"role" enums:"customer,vendor,admin" example:"customer"`
}

type VerifyOTPResponse struct {
	Verified    bool         `json:"verified" example:"true"`
	Phone       string       `json:"phone" example:"+989123456789"`
	AccessToken string       `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType   string       `json:"token_type" example:"Bearer"`
	ExpiresIn   int          `json:"expires_in" example:"3600"`
	User        UserResponse `json:"user"`
}

// MeResponse is what the token itself proves; no database lookup is needed.
type MeResponse struct {
	UserID string `json:"user_id" example:"5f0c6d1e-8b7a-4c1e-9a3f-2d4b6c8e0a12"`
	Role   string `json:"role" enums:"customer,vendor,admin" example:"customer"`
}
