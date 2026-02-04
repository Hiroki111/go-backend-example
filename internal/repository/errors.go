package repository

import "errors"

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrOptimisticLockFailed = errors.New("optimistic lock conflict")

var ErrItemNotFound = errors.New("item not found")
var ErrItemNotAvailable = errors.New("item not available")

var ErrProductAlreadyExists = errors.New("product already exists")
