package repository

import "errors"

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrOptimisticLockFailed = errors.New("optimistic lock conflict")

var ErrItemNotFound = errors.New("item not found")

var ErrProductAlreadyExists = errors.New("product already exists")
