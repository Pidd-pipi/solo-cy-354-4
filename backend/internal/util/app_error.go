// Package util provides shared helpers for the campus-market backend.
package util

import (
	"errors"
	"fmt"
)

// Sentinel errors returned by the repository layer.
var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrInvalidInput = errors.New("invalid input")
	ErrForbidden    = errors.New("forbidden")
	ErrUnauthorized = errors.New("unauthorized")
	ErrRateLimited  = errors.New("rate limited")
)

// AppError carries HTTP status, business code and a wrapped cause.
type AppError struct {
	Status  int
	Code    int
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Cause }

// NewAppError builds an AppError with an optional underlying cause.
func NewAppError(status, code int, message string, cause error) *AppError {
	return &AppError{Status: status, Code: code, Message: message, Cause: cause}
}

// WrapAppError rewraps err into an AppError keeping the chain alive.
func WrapAppError(err error, status, code int, message string) error {
	return &AppError{Status: status, Code: code, Message: message, Cause: err}
}
