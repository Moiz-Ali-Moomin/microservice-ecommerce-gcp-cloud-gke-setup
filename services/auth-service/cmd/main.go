package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"sync"

	"github.com/golang-jwt/jwt/v5"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/logger"
	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/tracing"
	"go.uber.org/zap"
)

var (
	jwtKey = []byte("my_secret_key")
	// In-memory user store: username -> password
	userStore = map[string]string{
		"admin": "password",
	}
	storeMutex sync.RWMutex
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

func main() {
	logger.Init("auth-service")

	tp, err := tracing.InitTracer("auth-service")
	if err != nil {
		logger.Log.Fatal("Failed to init tracer", zap.Error(err))
	}
	defer func() { _ = tp.Shutdown(nil) }()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/login", Login)
	mux.HandleFunc("/signup", Signup)
	mux.HandleFunc("/validate", Validate)

	logger.Log.Info("Starting auth-service on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func Signup(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	storeMutex.Lock()
	if _, exists := userStore[creds.Username]; exists {
		storeMutex.Unlock()
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}
	userStore[creds.Username] = creds.Password
	storeMutex.Unlock()

	w.WriteHeader(http.StatusCreated)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	storeMutex.RLock()
	expectedPassword, ok := userStore[creds.Username]
	storeMutex.RUnlock()

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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString, "username": creds.Username})
}

func Validate(w http.ResponseWriter, r *http.Request) {
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

