package errors

import "fmt"

type ErrorCode string

const (
	ErrCodeInvalidRequest    ErrorCode = "INVALID_REQUEST"
	ErrCodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden         ErrorCode = "FORBIDDEN"
	ErrCodeNotFound          ErrorCode = "NOT_FOUND"
	ErrCodeConflict          ErrorCode = "CONFLICT"
	ErrCodeRateLimited       ErrorCode = "RATE_LIMITED"
	ErrCodeTimeout           ErrorCode = "TIMEOUT"
	ErrCodeProviderError     ErrorCode = "PROVIDER_ERROR"
	ErrCodeNetworkError      ErrorCode = "NETWORK_ERROR"
	ErrCodeInsufficientFunds ErrorCode = "INSUFFICIENT_FUNDS"
	ErrCodeInvalidCard       ErrorCode = "INVALID_CARD"
	ErrCodeCardExpired       ErrorCode = "CARD_EXPIRED"
	ErrCodeCardDeclined      ErrorCode = "CARD_DECLINED"
	ErrCodeFraudRejection    ErrorCode = "FRAUD_REJECTION"
	ErrCodeInvalidToken      ErrorCode = "INVALID_TOKEN"
	ErrCodeDuplicatePayment  ErrorCode = "DUPLICATE_PAYMENT"
	ErrCodeInvalidAmount     ErrorCode = "INVALID_AMOUNT"
	ErrCodeUnsupportedMethod ErrorCode = "UNSUPPORTED_METHOD"
	ErrCodeShipmentCancelled ErrorCode = "SHIPMENT_CANCELLED"
	ErrCodeQRExpired         ErrorCode = "QR_EXPIRED"
	ErrCodePOSNotFound       ErrorCode = "POS_NOT_FOUND"
	ErrCodeInvalidWebhook    ErrorCode = "INVALID_WEBHOOK"
	ErrCodeInternal          ErrorCode = "INTERNAL_ERROR"
)

type SDKError struct {
	Code            ErrorCode
	Message         string
	ProviderCode    string
	ProviderMessage string
	Details         map[string]any
	Cause           error
}

func (e *SDKError) Error() string {
	if e.ProviderCode != "" {
		return fmt.Sprintf("[%s] %s (provider: %s - %s)", e.Code, e.Message, e.ProviderCode, e.ProviderMessage)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *SDKError) Unwrap() error {
	return e.Cause
}

func (e *SDKError) Is(target error) bool {
	if t, ok := target.(*SDKError); ok {
		return e.Code == t.Code
	}
	return false
}

func NewError(code ErrorCode, message string) *SDKError {
	return &SDKError{
		Code:    code,
		Message: message,
	}
}

func NewErrorWithCause(code ErrorCode, message string, cause error) *SDKError {
	return &SDKError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

func NewProviderError(code ErrorCode, message, providerCode, providerMessage string) *SDKError {
	return &SDKError{
		Code:            code,
		Message:         message,
		ProviderCode:    providerCode,
		ProviderMessage: providerMessage,
	}
}

func InvalidRequest(message string) *SDKError {
	return NewError(ErrCodeInvalidRequest, message)
}

func Unauthorized(message string) *SDKError {
	return NewError(ErrCodeUnauthorized, message)
}

func NotFound(resource string) *SDKError {
	return NewError(ErrCodeNotFound, fmt.Sprintf("%s not found", resource))
}

func ProviderError(providerCode, providerMessage string) *SDKError {
	return NewProviderError(ErrCodeProviderError, "provider error", providerCode, providerMessage)
}

func InsufficientFunds() *SDKError {
	return NewError(ErrCodeInsufficientFunds, "insufficient funds")
}

func InvalidCard(detail string) *SDKError {
	return NewError(ErrCodeInvalidCard, fmt.Sprintf("invalid card: %s", detail))
}

func RateLimited() *SDKError {
	return NewError(ErrCodeRateLimited, "rate limit exceeded")
}

func Timeout() *SDKError {
	return NewError(ErrCodeTimeout, "request timeout")
}

func IsNotFound(err error) bool {
	if e, ok := err.(*SDKError); ok {
		return e.Code == ErrCodeNotFound
	}
	return false
}

func IsUnauthorized(err error) bool {
	if e, ok := err.(*SDKError); ok {
		return e.Code == ErrCodeUnauthorized
	}
	return false
}

func IsRateLimited(err error) bool {
	if e, ok := err.(*SDKError); ok {
		return e.Code == ErrCodeRateLimited
	}
	return false
}

func IsProviderError(err error) bool {
	if e, ok := err.(*SDKError); ok {
		return e.Code == ErrCodeProviderError
	}
	return false
}
