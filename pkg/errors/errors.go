package errors

import (
	"fmt"
)

const (
	ErrCodeInternal    = "INTERNAL_ERROR"
	ErrCodeInvalidReq  = "INVALID_REQUEST"
	ErrCodeGeminiAPI   = "GEMINI_API_ERROR"
	ErrCodeImageGenAPI = "IMAGE_GEN_API_ERROR"
	ErrCodePPTRender   = "PPT_RENDER_ERROR"
	ErrCodeStorage     = "STORAGE_ERROR"
	ErrCodeRateLimited = "RATE_LIMITED"
	ErrCodeNotFound    = "NOT_FOUND"
)

type AppError struct {
	Code    string
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func New(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func Wrap(err error, code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

func Is(err error, code string) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == code
	}
	return false
}
