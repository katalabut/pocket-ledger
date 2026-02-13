package domain

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrValidation     = errors.New("validation error")
	ErrConflict       = errors.New("conflict")
	ErrSplitMismatch  = errors.New("split amounts do not match transaction amount")
	ErrForeignKey     = errors.New("referenced entity does not exist")
)
