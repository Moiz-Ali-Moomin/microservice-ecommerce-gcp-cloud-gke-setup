package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// Context keys
type contextKey string

const (
	ContextKeyUserID        contextKey = "user_id"
	ContextKeyCorrelationID contextKey = "correlation_id"
)

// EventContextMiddleware extracts headers and injects them into context
func EventContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract User ID
		userID := r.Header.Get("X-User-ID")
		if userID != "" {
			ctx = context.WithValue(ctx, ContextKeyUserID, userID)
		}

		// Extract Correlation ID or generate one
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}
		// Always set the header for downstream (if we were proxying, but here we just process)
		w.Header().Set("X-Correlation-ID", correlationID)

		ctx = context.WithValue(ctx, ContextKeyCorrelationID, correlationID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext helper
func GetUserIDFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(ContextKeyUserID).(string); ok {
		return val
	}
	return ""
}

// GetCorrelationIDFromContext helper
func GetCorrelationIDFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(ContextKeyCorrelationID).(string); ok {
		return val
	}
	return "" // Should ideally never verify empty if middleware is used
}
