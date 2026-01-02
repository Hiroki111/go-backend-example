package domain

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Name       string `gorm:"uniqueIndex;not null"`
	PriceCents int64  `gorm:"not null"`
}
