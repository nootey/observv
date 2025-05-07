package models

import (
	"time"
)

type User struct {
	Username      string     `gorm:"unique;not null" json:"username"`
	Password      string     `gorm:"not null" json:"-"` // do not output the password
	Email         string     `gorm:"unique;not null" json:"email"`
	EmailVerified *time.Time `json:"email_verified,omitempty"`
	Role          string     `gorm:"not null" json:"role"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
