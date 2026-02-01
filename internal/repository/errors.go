package repository

import "errors"

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrOptimisticLockFailed = errors.New("optimistic lock conflict")

var ErrItemNotFound = errors.New("item not found")

var ErrProductAlreadyExists = errors.New("product already exists")
