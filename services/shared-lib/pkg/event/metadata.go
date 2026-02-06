package event

import "time"

// Metadata contains standard fields required for all business events
type Metadata struct {
	EventID       string    `json:"event_id"`
	EventType     string    `json:"event_type"`
	ServiceName   string    `json:"service_name"`
	UserID        string    `json:"user_id,omitempty"`
	CorrelationID string    `json:"correlation_id"`
	Timestamp     time.Time `json:"timestamp"`
	Version       string    `json:"version"` // Schema version
}
