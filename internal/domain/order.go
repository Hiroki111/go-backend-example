package domain

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	ProductID  uint  `gorm:"not null"`
	UserID     uint  `gorm:"not null"`
	PriceCents int64 `gorm:"not null"`

	Product Product `gorm:"foreignKey:ProductID"`
	User    User    `gorm:"foreignKey:UserID"`
}
