package bootstrap

// Application wide constants. Anything that varies per environment belongs in
// env.go and must be read from an environment variable instead.
const (
	AppName = "too-good-to-go-backend"

	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"

	// HeaderRequestID is read from incoming requests (if a proxy already set
	// it) and always echoed back on the response.
	HeaderRequestID = "X-Request-ID"

	// Keys used with gin.Context.Set / Get.
	CtxKeyRequestID = "request_id"
	CtxKeyUserID    = "user_id" // set by the JWT auth middleware
	CtxKeyUserRole  = "user_role"

	HealthPath  = "/health"
	SwaggerPath = "/swagger"

	SMSProviderFake = "fake"
)
