package global_errors

import (
	"errors"
)

var (
	IncorrectHost = errors.New("incorrect host")
	InternalError = errors.New("internal error")
)
