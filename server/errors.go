package server

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidValue = errors.New("invalid value")
)
