package repository

import "errors"

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrItemNotFound = errors.New("item not found")
