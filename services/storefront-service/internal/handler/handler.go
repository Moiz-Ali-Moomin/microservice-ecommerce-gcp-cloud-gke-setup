package handler

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/storefront-service/internal/service"
)

// --- Shared Types ---

// Config holds service configuration
type Config struct {
	Port       string
	GatewayURL string
}

// Offer represents a product offer
type Offer struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Payout      float64 `json:"payout"` // Acting as Price for the demo
	VendorID    string  `json:"vendor_id"`
	Category    string  `json:"category"`
}

// CartItem represents an item in the cart
type CartItem struct {
	OfferID  string  `json:"offer_id"`
	VendorID string  `json:"vendor_id"`
	Price    float64 `json:"price"`
	Title    string  `json:"title"` // Enriched
	Quantity int     `json:"quantity"`
}

// Cart represents the user's cart
type Cart struct {
	UserID string     `json:"user_id"`
	Items  []CartItem `json:"items"`
}

// AuthResponse from API Gateway/Auth Service
type AuthResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

// --- Handler ---

type Handler struct {
	analytics *service.AnalyticsService
	config    Config
}

func NewHandler(analytics *service.AnalyticsService, config Config) *Handler {
	return &Handler{
		analytics: analytics,
		config:    config,
	}
}

// Home Handler
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	// Analytics
	h.analytics.TrackPageView(r.URL.Path, r.UserAgent())

	// Logic
	username := GetAuthenticatedUser(r)

	// Fetch offers from API Gateway
	resp, err := http.Get(h.config.GatewayURL + "/api/v1/offers")
	if err != nil {
		log.Printf("Error fetching offers: %v", err)
		http.Error(w, "Failed to fetch products", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	var offers []Offer
	if resp.StatusCode == http.StatusOK {
		json.NewDecoder(resp.Body).Decode(&offers)
	}

	data := map[string]interface{}{
		"Title":  "Home",
		"Offers": offers,
		"User":   username,
	}

	Render(w, "home.html", data)
}

// --- Helpers ---

var FuncMap = template.FuncMap{
	"mul": func(a float64, b int) float64 {
		return a * float64(b)
	},
}

func Render(w http.ResponseWriter, tmplName string, data interface{}) {
	tmpl, err := template.New("layout.html").Funcs(FuncMap).ParseFiles("templates/layout.html", "templates/"+tmplName)
	if err != nil {
		log.Printf("Template Parsing Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template Execution Error: %v", err)
	}
}

func GetAuthenticatedUser(r *http.Request) string {
	cookie, err := r.Cookie("auth_user")
	if err != nil {
		return ""
	}
	return cookie.Value
}
