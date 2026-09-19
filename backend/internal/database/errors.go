package database

import "errors"

// ErrDatabaseNotInitialized is returned when an operation is performed on an uninitialized database client.
var ErrDatabaseNotInitialized = errors.New("database client is not initialized")
