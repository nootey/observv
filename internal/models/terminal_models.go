package models

import (
	"time"
)

type Terminal struct {
	UserID    uint      `json:"user_id"`
	Status    string    `gorm:"type:varchar(100);enum(online,offline);not null" json:"status"`
	DataSent  int64     `json:"data_sent"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
