package dto

import "errors"

var ErrReplayDisabled = errors.New("relay event history is disabled")
var ErrSinkNotFound = errors.New("relay sink not found")
