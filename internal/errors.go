package internal

import "errors"

const (
	ValidationMsg = "validation error"
	SqlMsg        = "database error"
)

var (
	ErrValidation = errors.New(ValidationMsg)
	ErrSql        = errors.New(SqlMsg)
)
