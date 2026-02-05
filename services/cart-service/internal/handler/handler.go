package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/logger"
	"go.uber.org/zap"
)

type CartItem struct {
	OfferID  string `json:"offer_id"`
	Quantity int    `json:"quantity"`
}

type Cart struct {
	UserID string     `json:"user_id"`
	Items  []CartItem `json:"items"`
}

type Handler struct {
	carts map[string]*Cart
}

func New() *Handler {
	return &Handler{
		carts: make(map[string]*Cart),
	}
}

func (h *Handler) HandleCart(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "Missing user_id", http.StatusBadRequest)
		return
	}

	if h.carts[userID] == nil {
		h.carts[userID] = &Cart{
			UserID: userID,
			Items:  []CartItem{},
		}
	}

	if r.Method == http.MethodGet {
		json.NewEncoder(w).Encode(h.carts[userID])
		return
	}

	if r.Method == http.MethodPost {
		var item CartItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Check if item exists, update qty
		found := false
		for i, existing := range h.carts[userID].Items {
			if existing.OfferID == item.OfferID {
				h.carts[userID].Items[i].Quantity += item.Quantity
				found = true
				break
			}
		}
		if !found {
			h.carts[userID].Items = append(h.carts[userID].Items, item)
		}

		logger.Log.Info("Added to cart", zap.String("user_id", userID), zap.String("offer_id", item.OfferID))
		w.WriteHeader(http.StatusOK)
		return
	}

	// Handle Delete/Clear for Checkout
	if r.Method == http.MethodDelete {
		h.carts[userID].Items = []CartItem{}
		w.WriteHeader(http.StatusOK)
		return
	}
}

