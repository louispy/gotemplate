package custerr

import "errors"

var ErrDataNotFound = errors.New("data not found")
var ErrDuplicate = errors.New("duplicate data")
var ErrNoTransaction = errors.New("no open transaction")
