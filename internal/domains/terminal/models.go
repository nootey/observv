package terminal

import "time"

const (
	StatusOnline  = "online"
	StatusOffline = "offline"
)

type Terminal struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Status        string     `json:"status"`
	DataSentBytes int64      `json:"data_sent_bytes"`
	UptimeSeconds int64      `json:"uptime_seconds"`
	LastSeenAt    *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type HeartbeatRequest struct {
	Status string `json:"status"`
}

type TransferRequest struct {
	ToID  int64 `json:"to_id"`
	Bytes int64 `json:"bytes"`
}
