package listing

import "errors"

// ErrNotFound lets every layer recognize a missing listing with errors.Is.
var ErrNotFound = errors.New("listing not found")
