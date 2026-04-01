package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is the interface all application errors implement.
type AppError interface {
	error
	Code() string
	Message() string
	HTTPStatus() int
}

// As extracts the first AppError from an error chain.
func As(err error) (AppError, bool) {
	var appErr AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// --- NotFoundError ---

type NotFoundError struct {
	Resource string
}

func NewNotFound(resource string) *NotFoundError {
	return &NotFoundError{Resource: resource}
}

func (e *NotFoundError) Error() string      { return fmt.Sprintf("%s not found", e.Resource) }
func (e *NotFoundError) Code() string        { return "NOT_FOUND" }
func (e *NotFoundError) Message() string     { return e.Error() }
func (e *NotFoundError) HTTPStatus() int     { return http.StatusNotFound }

// --- InvalidInputError ---

type InvalidInputError struct {
	Msg    string
	Fields map[string]string
}

func NewInvalidInput(msg string, fields map[string]string) *InvalidInputError {
	return &InvalidInputError{Msg: msg, Fields: fields}
}

func (e *InvalidInputError) Error() string      { return e.Msg }
func (e *InvalidInputError) Code() string        { return "INVALID_INPUT" }
func (e *InvalidInputError) Message() string     { return e.Msg }
func (e *InvalidInputError) HTTPStatus() int     { return http.StatusBadRequest }

// --- ConflictError ---

type ConflictError struct {
	Msg  string
	Data any
}

func NewConflict(msg string, data any) *ConflictError {
	return &ConflictError{Msg: msg, Data: data}
}

func (e *ConflictError) Error() string      { return e.Msg }
func (e *ConflictError) Code() string        { return "CONFLICT" }
func (e *ConflictError) Message() string     { return e.Msg }
func (e *ConflictError) HTTPStatus() int     { return http.StatusConflict }

// --- UnauthorizedError ---

type UnauthorizedError struct {
	Msg string
}

func NewUnauthorized(msg string) *UnauthorizedError {
	return &UnauthorizedError{Msg: msg}
}

func (e *UnauthorizedError) Error() string      { return e.Msg }
func (e *UnauthorizedError) Code() string        { return "UNAUTHORIZED" }
func (e *UnauthorizedError) Message() string     { return e.Msg }
func (e *UnauthorizedError) HTTPStatus() int     { return http.StatusUnauthorized }

// --- ForbiddenError ---

type ForbiddenError struct {
	Msg string
}

func NewForbidden(msg string) *ForbiddenError {
	return &ForbiddenError{Msg: msg}
}

func (e *ForbiddenError) Error() string      { return e.Msg }
func (e *ForbiddenError) Code() string        { return "FORBIDDEN" }
func (e *ForbiddenError) Message() string     { return e.Msg }
func (e *ForbiddenError) HTTPStatus() int     { return http.StatusForbidden }

// --- InternalError ---

type InternalError struct {
	Msg string
	Err error // underlying error, never exposed to client
}

func NewInternal(msg string, err error) *InternalError {
	return &InternalError{Msg: msg, Err: err}
}

func (e *InternalError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}
	return e.Msg
}

func (e *InternalError) Code() string        { return "INTERNAL_ERROR" }
func (e *InternalError) Message() string     { return e.Msg }
func (e *InternalError) HTTPStatus() int     { return http.StatusInternalServerError }
func (e *InternalError) Unwrap() error       { return e.Err }
