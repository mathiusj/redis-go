package errors

import "fmt"

// RedisError represents a Redis protocol error
type RedisError struct {
	Code    string
	Message string
}

func (e RedisError) Error() string {
	return fmt.Sprintf("%s %s", e.Code, e.Message)
}

// Common errors
var (
	ErrWrongNumberOfArguments = RedisError{Code: "ERR", Message: "wrong number of arguments"}
	ErrUnknownCommand         = RedisError{Code: "ERR", Message: "unknown command"}
	ErrInvalidExpireTime      = RedisError{Code: "ERR", Message: "invalid expire time"}
	ErrSyntaxError            = RedisError{Code: "ERR", Message: "syntax error"}
	ErrUnsupportedParameter   = RedisError{Code: "ERR", Message: "unsupported CONFIG parameter"}
)

// WrongNumberOfArguments returns an error for incorrect argument count
func WrongNumberOfArguments(command string) RedisError {
	return RedisError{
		Code:    "ERR",
		Message: fmt.Sprintf("wrong number of arguments for '%s' command", command),
	}
}

// UnknownCommand returns an error for unknown commands
func UnknownCommand(command string) RedisError {
	return RedisError{
		Code:    "ERR",
		Message: fmt.Sprintf("unknown command '%s'", command),
	}
}

// InvalidExpireTime returns an error for invalid expiration times
func InvalidExpireTime(command string) RedisError {
	return RedisError{
		Code:    "ERR",
		Message: fmt.Sprintf("invalid expire time in %s", command),
	}
}
