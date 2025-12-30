package main

import "time"

// User represents a user in the system
type User struct {
	ID           int64
	Email        string
	Password     string
	ApplicationID *int64 // Optional: for multi-tenant support
	CreatedAt    time.Time
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	Token        string
	UserID       int64
	ApplicationID *int64 // Which application issued this token
	ExpiresAt    int64
	Revoked      bool
	CreatedAt    time.Time
}

// Application represents a registered application/client
type Application struct {
	ID                int64
	Name              string
	Domain            string
	APIKeyHash        string
	APIKeyPrefix      string
	RateLimitPerMinute int
	AllowedOrigins    []string
	Active            bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// Scope represents a permission scope
type Scope struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   time.Time
}

// TokenInfo represents token metadata for introspection
type TokenInfo struct {
	Active    bool
	UserID    *int64
	Scopes    []string
	ExpiresAt *int64
	ClientID  *string
}

