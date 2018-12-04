package aggregate

import "errors"

// error variables.
var (
	ErrInvalidFile  = errors.New("invalid file")
	ErrNotFoundFile = errors.New("no such a file")
)
