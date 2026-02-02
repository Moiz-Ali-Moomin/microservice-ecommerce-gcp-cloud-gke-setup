package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

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

// Auth Response from API Gateway/Auth Service
type AuthResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

// Global FuncMap
var funcMap = template.FuncMap{
	"mul": func(a float64, b int) float64 {
		return a * float64(b)
	},
}

func main() {
	config := Config{
		Port:       getEnv("PORT", "3000"),
		GatewayURL: getEnv("GATEWAY_URL", "http://api-gateway:8080"),
	}

	// Static Files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes
	http.HandleFunc("/", handleHome(config))
	http.HandleFunc("/login", handleLogin(config))
	http.HandleFunc("/signup", handleSignup(config))
	http.HandleFunc("/logout", handleLogout())

	// Protected Routes
	http.HandleFunc("/cart", handleCart(config))
	http.HandleFunc("/cart/add", handleAddToCart(config))
	http.HandleFunc("/cart/remove", handleRemoveFromCart(config))
	http.HandleFunc("/checkout", handleCheckout(config))
	http.HandleFunc("/payment/process", handlePaymentProcess(config))

	log.Printf("Storefront-Service (GO commerce) running on port %s", config.Port)
	log.Printf("Connected to Gateway at %s", config.GatewayURL)
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}

func render(w http.ResponseWriter, tmplName string, data interface{}) {
	tmpl, err := template.New("layout.html").Funcs(funcMap).ParseFiles("templates/layout.html", "templates/"+tmplName)
	if err != nil {
		log.Printf("Template Parsing Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template Execution Error: %v", err)
	}
}

// Helper to get authenticated user from cookie
func getAuthenticatedUser(r *http.Request) string {
	cookie, err := r.Cookie("auth_user")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func handleHome(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := getAuthenticatedUser(r)

		// Fetch offers from API Gateway
		resp, err := http.Get(cfg.GatewayURL + "/api/v1/offers")
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

		render(w, "home.html", data)
	}
}

func handleLogin(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			render(w, "login.html", map[string]interface{}{"Title": "Login"})
			return
		}

		// POST Login
		username := r.FormValue("username")
		password := r.FormValue("password")

		authBody, _ := json.Marshal(map[string]string{
			"username": username,
			"password": password,
		})

		resp, err := http.Post(cfg.GatewayURL+"/api/v1/login", "application/json", bytes.NewBuffer(authBody))
		if err != nil || resp.StatusCode != http.StatusOK {
			render(w, "login.html", map[string]interface{}{
				"Title": "Login",
				"Error": "Invalid credentials",
			})
			return
		}
		defer resp.Body.Close()

		var authResp AuthResponse
		json.NewDecoder(resp.Body).Decode(&authResp)

		// Set Cookies
		http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: authResp.Token, Path: "/", HttpOnly: true})
		http.SetCookie(w, &http.Cookie{Name: "auth_user", Value: authResp.Username, Path: "/"}) // Visible for UI

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func handleSignup(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			render(w, "signup.html", map[string]interface{}{"Title": "Create Account"})
			return
		}

		// POST Signup
		username := r.FormValue("username")
		password := r.FormValue("password")

		authBody, _ := json.Marshal(map[string]string{
			"username": username,
			"password": password,
		})

		resp, err := http.Post(cfg.GatewayURL+"/api/v1/signup", "application/json", bytes.NewBuffer(authBody))
		if err != nil || resp.StatusCode != http.StatusCreated {
			render(w, "signup.html", map[string]interface{}{
				"Title": "Create Account",
				"Error": "Failed to create account (User may exist)",
			})
			return
		}
		defer resp.Body.Close()

		// Auto-login or redirect to login
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: "", Path: "/", MaxAge: -1})
		http.SetCookie(w, &http.Cookie{Name: "auth_user", Value: "", Path: "/", MaxAge: -1})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func handleCart(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := getAuthenticatedUser(r)
		if userID == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// 1. Fetch Cart
		req, _ := http.NewRequest("GET", cfg.GatewayURL+"/api/v1/cart?user_id="+userID, nil)
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		var cart Cart
		if err == nil && resp.StatusCode == http.StatusOK {
			json.NewDecoder(resp.Body).Decode(&cart)
			resp.Body.Close()
		}

		// 2. Fetch Offers
		respOffers, err := http.Get(cfg.GatewayURL + "/api/v1/offers")
		offersMap := make(map[string]Offer)
		if err == nil && respOffers.StatusCode == http.StatusOK {
			var offers []Offer
			json.NewDecoder(respOffers.Body).Decode(&offers)
			respOffers.Body.Close()
			for _, o := range offers {
				offersMap[o.ID] = o
			}
		}

		// 3. Enrich Cart
		total := 0.0
		for i, item := range cart.Items {
			if offer, ok := offersMap[item.OfferID]; ok {
				cart.Items[i].Title = offer.Title
				cart.Items[i].Price = offer.Payout
				cart.Items[i].Quantity = item.Quantity
				total += offer.Payout * float64(item.Quantity)
			} else {
				cart.Items[i].Title = "Unknown Product"
			}
		}

		data := map[string]interface{}{
			"Title": "My Cart",
			"Cart":  cart,
			"Total": total,
			"User":  userID,
		}

		render(w, "cart.html", data)
	}
}

func handleCheckout(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := getAuthenticatedUser(r)
		if userID == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		req, _ := http.NewRequest("GET", cfg.GatewayURL+"/api/v1/cart?user_id="+userID, nil)
		client := &http.Client{Timeout: 5 * time.Second}
		resp, _ := client.Do(req)

		total := 0.0
		if resp != nil && resp.StatusCode == http.StatusOK {
			var cart Cart
			json.NewDecoder(resp.Body).Decode(&cart)
			resp.Body.Close()

			respOffers, _ := http.Get(cfg.GatewayURL + "/api/v1/offers")
			if respOffers != nil && respOffers.StatusCode == http.StatusOK {
				var offers []Offer
				json.NewDecoder(respOffers.Body).Decode(&offers)
				respOffers.Body.Close()
				offersMap := make(map[string]Offer)
				for _, o := range offers {
					offersMap[o.ID] = o
				}
				for _, item := range cart.Items {
					if off, ok := offersMap[item.OfferID]; ok {
						total += off.Payout * float64(item.Quantity)
					}
				}
			}
		}

		render(w, "payment.html", map[string]interface{}{
			"Title": "Payment Gateway",
			"Total": total,
			"User":  userID,
		})
	}
}

func handlePaymentProcess(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := getAuthenticatedUser(r)
		if userID == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if r.Method != "POST" {
			http.Redirect(w, r, "/payment", http.StatusSeeOther)
			return
		}

		// Call backend to clear cart
		req, _ := http.NewRequest("DELETE", cfg.GatewayURL+"/api/v1/cart?user_id="+userID, nil)
		client := &http.Client{Timeout: 5 * time.Second}
		client.Do(req)

		rand.Seed(time.Now().UnixNano())
		days := rand.Intn(5) + 3

		render(w, "checkout.html", map[string]interface{}{
			"Title":        "Order Confirmed",
			"DeliveryDays": days,
			"User":         userID,
		})
	}
}

func handleAddToCart(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := getAuthenticatedUser(r)
		if userID == "" {
			// If not logged in, redirect to login page
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		offerID := r.FormValue("offer_id")

		apiURL := fmt.Sprintf("%s/api/v1/cart?user_id=%s", cfg.GatewayURL, userID)

		item := map[string]interface{}{
			"offer_id": offerID,
			"quantity": 1,
		}
		jsonBody, _ := json.Marshal(item)

		resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			log.Printf("Failed to add to cart: %v", err)
		} else {
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				log.Printf("Cart service status: %d, body: %s", resp.StatusCode, string(body))
			}
			defer resp.Body.Close()
		}

		http.Redirect(w, r, "/cart", http.StatusSeeOther)
	}
}

func handleRemoveFromCart(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/cart", http.StatusSeeOther)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
