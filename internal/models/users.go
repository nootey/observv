package models

import (
	"time"
)

type User struct {
	Username      string     `gorm:"type:varchar(100);uniqueIndex:idx_username;not null" json:"username"`
	Password      string     `gorm:"not null" json:"-"`
	Email         string     `gorm:"type:varchar(255);uniqueIndex:idx_email;not null" json:"email"`
	EmailVerified *time.Time `json:"email_verified,omitempty"`
	Role          string     `gorm:"not null" json:"role"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
