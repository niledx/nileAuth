package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

// APIKeyAuth middleware validates API keys
func (a *App) APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health/ready endpoints
		if strings.HasPrefix(r.URL.Path, "/health") || strings.HasPrefix(r.URL.Path, "/ready") {
			next.ServeHTTP(w, r)
			return
		}

		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// Try Authorization header: Bearer <api-key>
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				apiKey = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if apiKey == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "API key required")
			return
		}

		// Validate API key by checking against all applications
		// In production, you'd want to optimize this with a cache or index
		app := a.validateAPIKey(apiKey)
		if app == nil || !app.Active {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid API key")
			return
		}

		// Store application in context
		ctx := context.WithValue(r.Context(), "application", app)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateAPIKey validates an API key by checking it against stored hashes
func (a *App) validateAPIKey(apiKey string) *Application {
	// Get prefix to narrow down candidates
	prefix := getAPIKeyPrefix(apiKey)

	// Get applications by prefix
	apps, err := a.DB.GetApplicationByAPIKeyPrefix(prefix)
	if err != nil || len(apps) == 0 {
		return nil
	}

	// Compare API key against each candidate using bcrypt
	for _, app := range apps {
		if err := bcrypt.CompareHashAndPassword([]byte(app.APIKeyHash), []byte(apiKey)); err == nil {
			return app
		}
	}

	return nil
}

// CORS middleware handles CORS headers
func (a *App) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get application from context if available
		var allowedOrigins []string
		if app, ok := r.Context().Value("application").(*Application); ok && app != nil {
			allowedOrigins = app.AllowedOrigins
		}

		origin := r.Header.Get("Origin")
		if origin != "" {
			// Check if origin is allowed
			allowed := false
			for _, o := range allowedOrigins {
				if o == origin || o == "*" {
					allowed = true
					break
				}
			}
			if allowed || len(allowedOrigins) == 0 {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RateLimiter implements per-application rate limiting
type RateLimiter struct {
	limiters map[int64]*rate.Limiter
	mu       sync.RWMutex
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiters: make(map[int64]*rate.Limiter),
	}
}

func (rl *RateLimiter) getLimiter(appID int64, limitPerMinute int) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[appID]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		// Double-check after acquiring write lock
		limiter, exists = rl.limiters[appID]
		if !exists {
			limiter = rate.NewLimiter(rate.Limit(limitPerMinute)/60, limitPerMinute)
			rl.limiters[appID] = limiter
		}
		rl.mu.Unlock()
	}

	return limiter
}

// RateLimit middleware enforces rate limits per application
func (a *App) RateLimit(next http.Handler) http.Handler {
	if a.rateLimiter == nil {
		a.rateLimiter = NewRateLimiter()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting for health/ready endpoints
		if strings.HasPrefix(r.URL.Path, "/health") || strings.HasPrefix(r.URL.Path, "/ready") {
			next.ServeHTTP(w, r)
			return
		}

		app, ok := r.Context().Value("application").(*Application)
		if !ok || app == nil {
			// No application context, allow through (will fail at API key auth)
			next.ServeHTTP(w, r)
			return
		}

		limiter := a.rateLimiter.getLimiter(app.ID, app.RateLimitPerMinute)
		if !limiter.Allow() {
			writeError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logging middleware logs requests
func (a *App) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		appID := "unknown"
		if app, ok := r.Context().Value("application").(*Application); ok && app != nil {
			appID = app.APIKeyPrefix
		}

		log.Printf("[%s] %s %s %d %v (app: %s)", r.Method, r.URL.Path, r.RemoteAddr, wrapped.statusCode, duration, appID)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// SecurityHeaders middleware adds security headers
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

// Helper functions for API key management
func generateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashAPIKey(apiKey string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	return string(hash), err
}

func getAPIKeyPrefix(apiKey string) string {
	if len(apiKey) >= 8 {
		return apiKey[:8]
	}
	return apiKey
}
