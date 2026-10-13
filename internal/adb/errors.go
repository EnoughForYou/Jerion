package adb

import "fmt"

// ErrorCode defines standardized ADB domain error classifications.
type ErrorCode string

const (
	ErrCodeAdbMissing           ErrorCode = "AdbMissing"
	ErrCodeDeviceUnauthorized   ErrorCode = "DeviceUnauthorized"
	ErrCodeDeviceOffline        ErrorCode = "DeviceOffline"
	ErrCodeTimeout              ErrorCode = "Timeout"
	ErrCodeParseFailure         ErrorCode = "ParseFailure"
	ErrCodePermissionDenied     ErrorCode = "PermissionDenied"
	ErrCodeNotSupportedOnThisOS ErrorCode = "NotSupportedOnThisOS"
	ErrCodeValidationFailed     ErrorCode = "ValidationFailed"
	ErrCodeCommandFailed        ErrorCode = "CommandFailed"
)

// ADBError encapsulates categorized ADB operational failures.
type ADBError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *ADBError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *ADBError) Unwrap() error {
	return e.Err
}

// NewADBError constructs a typed ADBError.
func NewADBError(code ErrorCode, message string, err error) *ADBError {
	return &ADBError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
