// Package kbx has default configuration values
package kbx

const (
	KeyringService        = "gnyx"
	DefaultKubexConfigDir = "$HOME/.gnyx"

	DefaultGNyxKeyPath    = "$HOME/.gnyxalize_be/github.com/kubex-ecosystem/gdnyx-key.pem"
	DefaultGNyxCertPath   = "$HOME/.gnyxalize_be/github.com/kubex-ecosystem/gdnyx-cert.pem"
	DefaultGNyxCAPath     = "$HOME/.gnyxalize_be/ca-cert.pem"
	DefaultGNyxConfigPath = "$HOME/.gnyxalize_be/config/config.json"

	DefaultConfigDir            = "$HOME/.gnyxalize_ds/config"
	DefaultConfigFile           = "$HOME/.gnyxalize_ds/config.json"
	DefaultKubexDomusConfigPath = "$HOME/.gnyxalize_ds/config/config.json"
)

const (
	DefaultVolumesDir     = "$HOME/.gnyxumes"
	DefaultRedisVolume    = "$HOME/.gnyxumes/redis"
	DefaultPostgresVolume = "$HOME/.gnyxumes/postgresql"
	DefaultMongoDBVolume  = "$HOME/.gnyxumes/mongodb"
	DefaultMongoVolume    = "$HOME/.gnyxumes/mongo"
	DefaultRabbitMQVolume = "$HOME/.gnyxumes/rabbitmq"
)

const (
	DefaultRateLimitLimit  = 100
	DefaultRateLimitBurst  = 100
	DefaultRequestWindow   = 1 * 60 * 1000 // 1 minute
	DefaultRateLimitJitter = 0.1
)

const (
	DefaultMaxRetries = 3
	DefaultRetryDelay = 1 * 1000 // 1 second
)

const (
	DefaultMaxIdleConns          = 100
	DefaultMaxIdleConnsPerHost   = 100
	DefaultIdleConnTimeout       = 90 * 1000 // 90 seconds
	DefaultTLSHandshakeTimeout   = 10 * 1000 // 10 seconds
	DefaultExpectContinueTimeout = 1 * 1000  // 1 second
	DefaultResponseHeaderTimeout = 5 * 1000  // 5 seconds
	DefaultTimeout               = 30 * 1000 // 30 seconds
	DefaultKeepAlive             = 30 * 1000 // 30 seconds
	DefaultMaxConnsPerHost       = 100
)

const (
	DefaultLLMProvider    = "gemini"
	DefaultLLMModel       = "gemini-2.0-flash"
	DefaultLLMMaxTokens   = 1024
	DefaultLLMTemperature = 0.3
)

const (
	DefaultApprovalRequireForResponses = false
	DefaultApprovalTimeoutMinutes      = 15
)

const (
	DefaultServerPort = "5000"
	DefaultServerHost = "0.0.0.0"
)

type ValidationError struct {
	Field   string
	Message string
}

func (v *ValidationError) Error() string {
	return v.Message
}
func (v *ValidationError) FieldError() map[string]string {
	return map[string]string{v.Field: v.Message}
}
func (v *ValidationError) FieldsError() map[string]string {
	return map[string]string{v.Field: v.Message}
}
func (v *ValidationError) ErrorOrNil() error {
	return v
}

var (
	ErrUsernameRequired = &ValidationError{Field: "username", Message: "Username is required"}
	ErrPasswordRequired = &ValidationError{Field: "password", Message: "Password is required"}
	ErrEmailRequired    = &ValidationError{Field: "email", Message: "Email is required"}
	ErrDBNotProvided    = &ValidationError{Field: "db", Message: "Database not provided"}
	ErrModelNotFound    = &ValidationError{Field: "model", Message: "Model not found"}
)
