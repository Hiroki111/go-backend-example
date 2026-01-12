package domain

import "gorm.io/gorm"

const (
	AdminRole    = "admin"
	CustomerRole = "customer"
)

type User struct {
	gorm.Model
	UserName string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
	Role     string `gorm:"not null"`
}
