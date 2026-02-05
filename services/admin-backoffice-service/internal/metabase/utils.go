package metabase

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SignMetabaseToken creates a signed JWT for embedding a specific dashboard.
// secretKey: The MB_ENCRYPTION_SECRET_KEY from your Metabase instance.
// dashboardID: The ID of the dashboard to embed.
// params: Optional parameters to lock the dashboard to specific data (e.g., {"merchant_id": 123}).
func SignMetabaseToken(dashboardID int, params map[string]interface{}) (string, error) {
	secretKey := os.Getenv("METABASE_SECRET_KEY")
	if secretKey == "" {
		return "", fmt.Errorf("METABASE_SECRET_KEY environment variable is not set")
	}

	// Metabase expects the claim "resource" to be an object like {"dashboard": 1}
	claims := jwt.MapClaims{
		"resource": map[string]int{"dashboard": dashboardID},
		"params":   params,
		"exp":      time.Now().Add(time.Minute * 10).Unix(), // Token valid for 10 minutes
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}
