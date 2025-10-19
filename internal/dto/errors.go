package dto

import (
	"errors"
)

var (
	ErrNotFound      = errors.New("errRecordNotFound")
	ErrAlreadyExists = errors.New("errAlreadyExists")
)
