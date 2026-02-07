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

// Event represents a generic event wrapper
type Event struct {
	EventID   string                 `json:"event_id"`
	Timestamp time.Time              `json:"timestamp"`
	Service   string                 `json:"service"`
	Metadata  map[string]interface{} `json:"metadata"`
	SessionID string                 `json:"session_id,omitempty"`
	OfferID   string                 `json:"offer_id,omitempty"`
}
