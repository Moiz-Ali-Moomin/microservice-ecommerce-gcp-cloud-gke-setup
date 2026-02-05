package metabase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// TokenResponse is the JSON response sent to the frontend
type TokenResponse struct {
	EmbedUrl string `json:"embed_url"`
	Token    string `json:"token"`
}

// GetAnalyticsTokenHandler handles the /api/analytics/token request.
// In a real app, this would inspect the user's session to determine which data they can see.
func GetAnalyticsTokenHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get Authentication (Mocked for this example)
	// userID := r.Context().Value("user_id")

	// 2. Define Parameters (Row-Level Security)
	// We want to lock this view to merchant_id = 99 (for example)
	// params := map[string]interface{}{"merchant_id": 99}
	params := map[string]interface{}{} // Empty for now

	// 3. Define the Dashboard ID to show (Configurable or hardcoded)
	// In production, look this up from config based on the user's role.
	dashboardID := 1

	// 4. Sign the Token
	token, err := SignMetabaseToken(dashboardID, params)
	if err != nil {
		http.Error(w, "Failed to sign token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Construct the full Embed URL
	siteUrl := os.Getenv("METABASE_SITE_URL")
	if siteUrl == "" {
		http.Error(w, "METABASE_SITE_URL not configured", http.StatusInternalServerError)
		return
	}

	// iframe URL format: SITE_URL/embed/dashboard/TOKEN#bordered=true&titled=true
	embedUrl := fmt.Sprintf("%s/embed/dashboard/%s#bordered=true&titled=true", siteUrl, token)

	// 6. Return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TokenResponse{
		EmbedUrl: embedUrl,
		Token:    token,
	})
}
