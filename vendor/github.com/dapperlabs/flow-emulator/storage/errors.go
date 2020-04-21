package storage

import "errors"

// ErrNotFound is an error returned when an entity cannot be found.
var ErrNotFound = errors.New("could not find entity")
