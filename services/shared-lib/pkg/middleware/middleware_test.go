package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/middleware"
)

func TestChain_AppliesMiddlewaresInOrder(t *testing.T) {
	order := []string{}

	a := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "a-before")
			next.ServeHTTP(w, r)
			order = append(order, "a-after")
		})
	}
	b := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "b-before")
			next.ServeHTTP(w, r)
			order = append(order, "b-after")
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	chained := middleware.Chain(handler, a, b)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	chained.ServeHTTP(rr, req)

	expected := []string{"b-before", "a-before", "handler", "a-after", "b-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected order %v, got %v", expected, order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("step %d: expected %q, got %q", i, v, order[i])
		}
	}
}

func TestMetricsMiddleware_Returns200(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok")) //nolint:errcheck
	})

	chained := middleware.Chain(handler, middleware.Metrics)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	rr := httptest.NewRecorder()
	chained.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestMetricsMiddleware_Captures4xxStatus(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	chained := middleware.Chain(handler, middleware.Metrics)
	req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
	rr := httptest.NewRecorder()
	chained.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestRateLimiter_AllowsUnderBurst(t *testing.T) {
	rl := middleware.NewRateLimiter(100, 20)
	handler := rl.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// 20 requests should all succeed within burst
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i+1, rr.Code)
		}
	}
}

func TestRateLimiter_Blocks429WhenExceedBurst(t *testing.T) {
	// 1 request/s, burst 1 → second request should be blocked
	rl := middleware.NewRateLimiter(1, 1)
	handler := rl.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ip := "192.168.1.100:9999"

	// First request → should pass
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = ip
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Errorf("first request: expected 200, got %d", rr1.Code)
	}

	// Second request immediately → burst exhausted → should 429
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = ip
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: expected 429, got %d", rr2.Code)
	}
}

func TestRequestIDMiddleware_InjectsID(t *testing.T) {
	var capturedID string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = r.Header.Get("X-Request-ID")
		w.WriteHeader(http.StatusOK)
	})

	chained := middleware.Chain(handler, middleware.RequestID)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	chained.ServeHTTP(rr, req)

	if capturedID == "" {
		t.Error("expected X-Request-ID to be set on request, got empty string")
	}
	if rr.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID to be set on response header")
	}
}

func TestRequestIDMiddleware_PreservesExistingID(t *testing.T) {
	const existingID = "my-upstream-request-id"
	var capturedID string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = r.Header.Get("X-Request-ID")
		w.WriteHeader(http.StatusOK)
	})

	chained := middleware.Chain(handler, middleware.RequestID)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", existingID)
	rr := httptest.NewRecorder()
	chained.ServeHTTP(rr, req)

	if capturedID != existingID {
		t.Errorf("expected X-Request-ID=%q, got %q", existingID, capturedID)
	}
}
