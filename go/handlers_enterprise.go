package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// HandleTokenIntrospect implements OAuth 2.0 token introspection
// POST /api/v1/auth/introspect
func (a *App) HandleTokenIntrospect(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Token == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Token is required")
		return
	}

	// Try to parse as JWT access token first
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	info := TokenInfo{Active: false}

	if err == nil && token.Valid {
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok {
			info.Active = true
			if userId, ok := claims["userId"].(float64); ok {
				uid := int64(userId)
				info.UserID = &uid
			}
			if exp, ok := claims["exp"].(float64); ok {
				expTime := int64(exp)
				info.ExpiresAt = &expTime
			}
		}
	} else {
		// Try as refresh token
		rt, _ := a.DB.GetRefreshToken(req.Token)
		if rt != nil && !rt.Revoked && rt.ExpiresAt > time.Now().Unix() {
			info.Active = true
			info.UserID = &rt.UserID
			info.ExpiresAt = &rt.ExpiresAt
		}
	}

	writeJSON(w, http.StatusOK, info)
}

// HandleTokenValidate validates an access token
// GET /api/v1/auth/validate?token=...
func (a *App) HandleTokenValidate(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		// Try Authorization header
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			tokenStr = strings.TrimPrefix(auth, "Bearer ")
		}
	}

	if tokenStr == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Token is required")
		return
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Token is invalid or expired")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid token claims")
		return
	}

	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"valid":  true,
		"userId": claims["userId"],
		"exp":    claims["exp"],
	})
}

// HandleCreateApplication creates a new application/client
// POST /api/v1/admin/applications
func (a *App) HandleCreateApplication(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name              string   `json:"name"`
		Domain            string   `json:"domain"`
		RateLimitPerMinute int     `json:"rate_limit_per_minute"`
		AllowedOrigins    []string `json:"allowed_origins"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Name == "" || req.Domain == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Name and domain are required")
		return
	}

	if req.RateLimitPerMinute <= 0 {
		req.RateLimitPerMinute = 100 // default
	}

	// Generate API key
	apiKey, err := generateAPIKey()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate API key")
		return
	}

	apiKeyHash, err := hashAPIKey(apiKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to hash API key")
		return
	}

	apiKeyPrefix := getAPIKeyPrefix(apiKey)

	app, err := a.DB.CreateApplication(req.Name, req.Domain, apiKeyHash, apiKeyPrefix, req.RateLimitPerMinute, req.AllowedOrigins)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create application")
		return
	}

	// Return API key only once (should be stored securely by client)
	writeSuccess(w, http.StatusCreated, map[string]interface{}{
		"application": map[string]interface{}{
			"id":                  app.ID,
			"name":                app.Name,
			"domain":              app.Domain,
			"api_key_prefix":      app.APIKeyPrefix,
			"rate_limit_per_minute": app.RateLimitPerMinute,
			"allowed_origins":     app.AllowedOrigins,
		},
		"api_key": apiKey, // Only returned on creation
	})
}

// HandleGetApplications lists all applications
// GET /api/v1/admin/applications
func (a *App) HandleGetApplications(w http.ResponseWriter, r *http.Request) {
	// This would require a ListApplications method in DB interface
	// For now, return not implemented
	writeError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "List applications not yet implemented")
}

// HandleRevokeToken revokes a specific token
// POST /api/v1/auth/revoke
func (a *App) HandleRevokeToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Token == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Token is required")
		return
	}

	// Try to revoke as refresh token
	err := a.DB.RevokeRefreshToken(req.Token)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_TOKEN", "Token not found or already revoked")
		return
	}

	writeSuccess(w, http.StatusOK, map[string]bool{"revoked": true})
}

