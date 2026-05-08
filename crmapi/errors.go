package crmapi

import "fmt"

// SDKError is a marker interface for all SDK-specific errors.
type SDKError interface {
	error
	isSDKError()
}

// ConfigError reports invalid SDK configuration.
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}

func (e *ConfigError) isSDKError() {}

// ValidationError reports invalid request input.
//
// Code is the upstream error_code surfaced by CRM for HTTP 400/422
// responses (e.g. "invalid_token"). Empty when the validation came
// from the SDK itself (client-side checks).
type ValidationError struct {
	Message string
	Code    string
}

func (e *ValidationError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s (code=%s)", e.Message, e.Code)
	}
	return e.Message
}

func (e *ValidationError) isSDKError() {}

// AuthError reports authentication or authorization failures.
type AuthError struct {
	Message string
	Code    string
	Status  int
}

func (e *AuthError) Error() string {
	if e.Code != "" && e.Status > 0 {
		return fmt.Sprintf("%s (code=%s, status=%d)", e.Message, e.Code, e.Status)
	}
	if e.Code != "" {
		return fmt.Sprintf("%s (code=%s)", e.Message, e.Code)
	}
	if e.Status > 0 {
		return fmt.Sprintf("%s (status=%d)", e.Message, e.Status)
	}
	return e.Message
}

func (e *AuthError) isSDKError() {}

// APIError reports a non-auth application-level CRM API error.
type APIError struct {
	Message string
	Code    string
	Status  int
}

func (e *APIError) Error() string {
	if e.Code != "" && e.Status > 0 {
		return fmt.Sprintf("%s (code=%s, status=%d)", e.Message, e.Code, e.Status)
	}
	if e.Code != "" {
		return fmt.Sprintf("%s (code=%s)", e.Message, e.Code)
	}
	if e.Status > 0 {
		return fmt.Sprintf("%s (status=%d)", e.Message, e.Status)
	}
	return e.Message
}

func (e *APIError) isSDKError() {}

// HTTPError reports transport or response decoding failures.
type HTTPError struct {
	Message string
	Status  int
	Cause   error
}

func (e *HTTPError) Error() string {
	switch {
	case e.Cause != nil && e.Status > 0:
		return fmt.Sprintf("%s (status=%d): %v", e.Message, e.Status, e.Cause)
	case e.Cause != nil:
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	case e.Status > 0:
		return fmt.Sprintf("%s (status=%d)", e.Message, e.Status)
	default:
		return e.Message
	}
}

func (e *HTTPError) Unwrap() error {
	return e.Cause
}

func (e *HTTPError) isSDKError() {}
