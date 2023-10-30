package requests

import "github.com/pkg/errors"

var (
	ErrRequestBodyDecodingFailed = errors.New("failed to decode a request body")
	ErrEmptyRequestField         = errors.New("field must not be empty")
)
