package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"sync"

	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/event"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/logger"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/tracing"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	jwtKey = []byte("my_secret_key")
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type Handler struct {
	userStore  map[string]string
	storeMutex *sync.RWMutex
	producer   *event.Producer
}

func NewHandler(p *event.Producer) *Handler {
	return &Handler{
		userStore: map[string]string{
			"admin": "password",
		},
		storeMutex: &sync.RWMutex{},
		producer:   p,
	}
}

func main() {
	logger.Init("auth-service")

	tp, err := tracing.InitTracer("auth-service")
	if err != nil {
		logger.Log.Fatal("Failed to init tracer", zap.Error(err))
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	// Init Producer
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"kafka:9092"}
	}
	producer, err := event.NewProducer(brokers, "auth-service")
	if err != nil {
		logger.Log.Fatal("Failed to create kafka producer", zap.Error(err))
	}
	defer producer.Close()

	h := NewHandler(producer)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/login", h.Login)
	mux.HandleFunc("/signup", h.Signup)
	mux.HandleFunc("/validate", h.Validate)

	logger.Log.Info("Starting auth-service on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.storeMutex.Lock()
	if _, exists := h.userStore[creds.Username]; exists {
		h.storeMutex.Unlock()
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}
	h.userStore[creds.Username] = creds.Password
	h.storeMutex.Unlock()

	// Emit user.signup
	evt := event.Event{
		EventID:   uuid.New().String(),
		Timestamp: time.Now(),
		Service:   "auth-service",
		Metadata: map[string]interface{}{
			"username": creds.Username,
			"action":   "signup",
		},
	}
	if err := h.producer.Emit(r.Context(), "user.signup", creds.Username, evt); err != nil {
		logger.Log.Error("Failed to emit signup event", zap.Error(err))
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.storeMutex.RLock()
	expectedPassword, ok := h.userStore[creds.Username]
	h.storeMutex.RUnlock()

	if !ok || expectedPassword != creds.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(60 * time.Minute)
	claims := &Claims{
		Username: creds.Username,
		Role:     "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Emit user.login
	evt := event.Event{
		EventID:   uuid.New().String(),
		Timestamp: time.Now(),
		Service:   "auth-service",
		Metadata: map[string]interface{}{
			"username": creds.Username,
			"action":   "login",
		},
	}
	if err := h.producer.Emit(r.Context(), "user.login", creds.Username, evt); err != nil {
		logger.Log.Error("Failed to emit login event", zap.Error(err))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString, "username": creds.Username})
}

func (h *Handler) Validate(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.Header.Get("Authorization")
	// Strip "Bearer " if present
	if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
		tokenStr = tokenStr[7:]
	}

	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusOK)
}
