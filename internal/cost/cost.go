package cost

import "time"

// Entry represents a single cost record for a Claude session.
type Entry struct {
	SessionID string    `json:"session_id"`
	Cost      float64   `json:"cost"`
	Timestamp time.Time `json:"timestamp"`
}
