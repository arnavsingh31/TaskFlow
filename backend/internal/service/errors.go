package service

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("email already exists")
	ErrUnauthorized = errors.New("unauthorized")
)
