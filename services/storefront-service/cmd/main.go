package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/event"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/logger"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/storefront-service/internal/handler"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/storefront-service/internal/service"
)

func main() {
	// Initialize Logger
	logger.Init("storefront-service")

	config := handler.Config{
		Port:       getEnv("PORT", "3000"),
		GatewayURL: getEnv("GATEWAY_URL", "http://api-gateway:8080"),
	}

	// Initialize Kafka Producer
	brokers := strings.Split(getEnv("KAFKA_BROKERS", "kafka:9092"), ",")
	producer, err := event.NewProducer(brokers, "page.viewed")
	if err != nil {
		log.Printf("Warning: Failed to initialize Kafka producer: %v", err)
		// Proceed without Kafka (producer will be nil)
	} else {
		defer producer.Close()
		log.Printf("Connected to Kafka brokers at %s", brokers)
	}

	// Initialize Services & Handler
	analyticsService := service.NewAnalyticsService(producer)
	h := handler.NewHandler(analyticsService, config)

	// Static Files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes
	http.HandleFunc("/", h.Home) // Use new handler
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

func handleLogin(cfg handler.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			handler.Render(w, "login.html", map[string]interface{}{"Title": "Login"})
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
			handler.Render(w, "login.html", map[string]interface{}{
				"Title": "Login",
				"Error": "Invalid credentials",
			})
			return
		}
		defer resp.Body.Close()

		var authResp handler.AuthResponse
		json.NewDecoder(resp.Body).Decode(&authResp)

		// Set Cookies
		http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: authResp.Token, Path: "/", HttpOnly: true})
		http.SetCookie(w, &http.Cookie{Name: "auth_user", Value: authResp.Username, Path: "/"}) // Visible for UI

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func handleSignup(cfg handler.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			handler.Render(w, "signup.html", map[string]interface{}{"Title": "Create Account"})
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
			handler.Render(w, "signup.html", map[string]interface{}{
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

func handleCart(cfg handler.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := handler.GetAuthenticatedUser(r)
		if userID == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// 1. Fetch Cart
		req, _ := http.NewRequest("GET", cfg.GatewayURL+"/api/v1/cart?user_id="+userID, nil)
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)

		var cart handler.Cart
		if err == nil && resp.StatusCode == http.StatusOK {
			json.NewDecoder(resp.Body).Decode(&cart)
			resp.Body.Close()
		}

		// 2. Fetch Offers
		respOffers, err := http.Get(cfg.GatewayURL + "/api/v1/offers")
		offersMap := make(map[string]handler.Offer)
		if err == nil && respOffers.StatusCode == http.StatusOK {
			var offers []handler.Offer
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

		handler.Render(w, "cart.html", data)
	}
}

func handleCheckout(cfg handler.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := handler.GetAuthenticatedUser(r)
		if userID == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		req, _ := http.NewRequest("GET", cfg.GatewayURL+"/api/v1/cart?user_id="+userID, nil)
		client := &http.Client{Timeout: 5 * time.Second}
		resp, _ := client.Do(req)

		total := 0.0
		if resp != nil && resp.StatusCode == http.StatusOK {
			var cart handler.Cart
			json.NewDecoder(resp.Body).Decode(&cart)
			resp.Body.Close()

			respOffers, _ := http.Get(cfg.GatewayURL + "/api/v1/offers")
			if respOffers != nil && respOffers.StatusCode == http.StatusOK {
				var offers []handler.Offer
				json.NewDecoder(respOffers.Body).Decode(&offers)
				respOffers.Body.Close()
				offersMap := make(map[string]handler.Offer)
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

		handler.Render(w, "payment.html", map[string]interface{}{
			"Title": "Payment Gateway",
			"Total": total,
			"User":  userID,
		})
	}
}

func handlePaymentProcess(cfg handler.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := handler.GetAuthenticatedUser(r)
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

		handler.Render(w, "checkout.html", map[string]interface{}{
			"Title":        "Order Confirmed",
			"DeliveryDays": days,
			"User":         userID,
		})
	}
}

func handleAddToCart(cfg handler.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := handler.GetAuthenticatedUser(r)
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

func handleRemoveFromCart(cfg handler.Config) http.HandlerFunc {
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
