package scrpc

import "errors"

var (
	ErrThrottled = errors.New("request is throttled")
)
