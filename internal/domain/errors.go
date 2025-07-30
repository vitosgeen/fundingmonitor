package domain

import "errors"

var (
	ErrExchangeNotFound = errors.New("exchange not found")
	ErrInvalidConfig    = errors.New("invalid configuration")
	ErrLogFileNotFound  = errors.New("log file not found")
) 