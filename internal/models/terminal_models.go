package models

import (
	"time"
)

type Terminal struct {
	ID            uint
	UserID        uint      `json:"user_id"`
	Status        string    `gorm:"type:varchar(100);enum(online,offline);not null" json:"status"`
	DataSent      int64     `json:"data_sent"`
	UptimeSeconds int64     `gorm:"default:0" json:"uptime_seconds"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
