package lwm2m

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrEmpty    = errors.New("empty")

	ErrCoreLinkInvalidValue = errors.New("invalid core link string value")

	ErrPathNotMatch = errors.New("wrong path type")
	ErrNodeNotFound = errors.New("node not found")
)
