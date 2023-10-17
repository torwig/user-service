package entities

import "github.com/pkg/errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserDeleted  = errors.New("user was deleted")
)
