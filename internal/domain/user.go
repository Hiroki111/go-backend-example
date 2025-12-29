package domain

import "gorm.io/gorm"

type User struct {
	gorm.Model
	UserName string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
}
