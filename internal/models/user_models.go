package models

import (
	"time"
)

type User struct {
	ID            uint
	Username      string     `gorm:"type:varchar(100);unique;not null" json:"username"`
	Password      string     `gorm:"not null" json:"-"`
	Email         string     `gorm:"type:varchar(255);unique;not null" json:"email"`
	EmailVerified *time.Time `json:"email_verified,omitempty"`
	Role          string     `gorm:"not null" json:"role"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
