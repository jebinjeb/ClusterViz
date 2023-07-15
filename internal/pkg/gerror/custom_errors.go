package gerror

// ErrorCode service gerror constants
type ErrorCode string

const (
	InternalError        ErrorCode = "Internal Error"
	ServiceSetup         ErrorCode = "ServiceSetup"
	ValidationFailed     ErrorCode = "Validations Failed"
	BadRequest           ErrorCode = "Bad Request"
	NotFound             ErrorCode = "Not Found"
	LDAPClient           ErrorCode = "LDAP Client"
	TokenNotFound        ErrorCode = "Token NotFound"
	AuthenticationFailed ErrorCode = "Authentication Failed"

	InvalidDBConfig ErrorCode = "Invalid DB Configurations"
	InvalidInput    ErrorCode = "Bad Request"
)
