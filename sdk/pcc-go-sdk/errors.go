package pccclient

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidBaseURL      = errors.New("pccclient: invalid base URL")
	ErrCredentialsRequired = errors.New("pccclient: access key and secret are required")
	ErrNamespaceRequired   = errors.New("pccclient: namespace is required")
	ErrConfigRequired      = errors.New("pccclient: config name is required")
	ErrTargetRequired      = errors.New("pccclient: target must be a non-nil pointer")
	ErrFieldNotFound       = errors.New("pccclient: field not found")
	ErrUnsupportedFormat   = errors.New("pccclient: unsupported config format")
)

type HTTPError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *HTTPError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("pccclient: server returned %s", e.Status)
	}
	return fmt.Sprintf("pccclient: server returned %s: %s", e.Status, e.Body)
}

type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("pccclient: API error %d: %s", e.Code, e.Message)
}

type DecodeError struct {
	ConfigName string
	Format     int
	Err        error
}

func (e *DecodeError) Error() string {
	return fmt.Sprintf(
		"pccclient: decode config %q (format %d): %v",
		e.ConfigName,
		e.Format,
		e.Err,
	)
}

func (e *DecodeError) Unwrap() error {
	return e.Err
}
