package data

import (
	"time"
)

type VerificationToken struct {
	ID         uint `gorm:"primary_key"`
	Token      string
	ExpiryDate time.Time
	UserID     string `gorm:"column:user_id"`
}
