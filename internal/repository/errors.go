package repository

import "errors"

// ErrNotFound is returned when a requested record does not exist. The service
// layer translates this into a domain-level not-found error.
var ErrNotFound = errors.New("record not found")
