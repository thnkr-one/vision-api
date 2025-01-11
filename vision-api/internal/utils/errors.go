package utils

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidInput indicates invalid input parameters
	ErrInvalidInput = errors.New("invalid input")

	// ErrProcessingFailed indicates a processing operation failed
	ErrProcessingFailed = errors.New("processing failed")

	// ErrRateLimitExceeded indicates API rate limit was exceeded
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrImageTooLarge indicates the image size exceeds the maximum allowed
	ErrImageTooLarge = errors.New("image too large")

	// ErrUnsupportedFormat indicates an unsupported image format
	ErrUnsupportedFormat = errors.New("unsupported image format")

	// ErrTimeout indicates an operation timed out
	ErrTimeout = errors.New("operation timed out")
)

// ProcessError represents a detailed processing error
type ProcessError struct {
	Op      string // Operation that failed
	File    string // File being processed
	Err     error  // Underlying error
	Details string // Additional error details
}

// Error implements the error interface
func (e *ProcessError) Error() string {
	if e.File == "" {
		return fmt.Sprintf("operation %s failed: %v - %s", e.Op, e.Err, e.Details)
	}
	return fmt.Sprintf("operation %s failed for %s: %v - %s", e.Op, e.File, e.Err, e.Details)
}

// Unwrap implements error unwrapping
func (e *ProcessError) Unwrap() error {
	return e.Err
}

// NewProcessError creates a new process error
func NewProcessError(op, file string, err error, details string) *ProcessError {
	return &ProcessError{
		Op:      op,
		File:    file,
		Err:     err,
		Details: details,
	}
}

// IsRateLimitError checks if an error is a rate limit error
func IsRateLimitError(err error) bool {
	return errors.Is(err, ErrRateLimitExceeded)
}

// IsTimeout checks if an error is a timeout error
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsInvalidInput checks if an error is an invalid input error
func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// WrapError wraps an error with additional context
func WrapError(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

// ExtractErrorDetails extracts structured details from an error
func ExtractErrorDetails(err error) map[string]interface{} {
	details := make(map[string]interface{})

	var processErr *ProcessError
	if errors.As(err, &processErr) {
		details["operation"] = processErr.Op
		details["file"] = processErr.File
		details["details"] = processErr.Details
		details["error"] = processErr.Err.Error()
		return details
	}

	details["error"] = err.Error()
	return details
}

// ErrorResponse represents an error response structure
type ErrorResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(code int, err error) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: err.Error(),
		Details: ExtractErrorDetails(err),
	}
}