package constants

// Unified business error codes used by the {code,message,data} response envelope.
const (
	CodeOK            = 0
	CodeBadRequest    = 40000
	CodeUnauthorized  = 40100
	CodeForbidden     = 40300
	CodeNotFound      = 40400
	CodeConflict      = 40900
	CodeRateLimited   = 42900
	CodeValidation    = 42200
	CodeInternalError = 50000
)
