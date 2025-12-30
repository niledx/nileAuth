package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cfg "github.com/example/nileauth/internal/config"
	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

var jwtSecret []byte

type App struct {
	DB          DB
	rateLimiter *RateLimiter
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}

func main() {
	c, err := cfg.New()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	jwtSecret = []byte(c.JwtSecret)

	var db DB
	switch c.DBAdapter {
	case "sqlite":
		s, err := NewSQLiteDB(c.SQLiteFile)
		if err != nil {
			log.Fatalf("sqlite init: %v", err)
		}
		db = s
	case "postgres":
		dsn, err := c.BuildPostgresDSN()
		if err != nil {
			log.Fatalf("postgres config error: %v", err)
		}
		
		// Apply migrations before connecting
		log.Println("Applying database migrations...")
		if err := ApplyMigrations("./migrations", dsn); err != nil {
			log.Printf("migrations warning: %v", err)
			// Don't fail if migrations already applied
			if err.Error() != "no change" {
				log.Printf("Migration error (continuing anyway): %v", err)
			}
		} else {
			log.Println("Migrations applied successfully")
		}
		
		p, err := NewPostgresDB(dsn)
		if err != nil {
			log.Fatalf("postgres init: %v", err)
		}
		db = p
		log.Println("Connected to PostgreSQL database")
	case "memory":
		log.Println("Using in-memory database (not recommended for production)")
		db = NewMemoryDB()
	default:
		log.Fatalf("unsupported DB_ADAPTER: %s (supported: postgres, sqlite, memory)", c.DBAdapter)
	}

	app := &App{DB: db}
	r := mux.NewRouter()

	// Apply global middleware
	r.Use(SecurityHeaders)
	r.Use(app.Logging)
	r.Use(app.CORS)

	// Health check endpoints (no auth required)
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")
	r.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if p, ok := app.DB.(interface{ ping() bool }); ok {
			if !p.ping() {
				w.WriteHeader(503)
				w.Write([]byte(`{"ready":false}`))
				return
			}
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ready":true}`))
	}).Methods("GET")

	// API v1 routes with authentication and rate limiting
	v1 := r.PathPrefix("/api/v1").Subrouter()
	v1.Use(app.APIKeyAuth)
	v1.Use(app.RateLimit)

	// Authentication endpoints
	v1.HandleFunc("/auth/register", app.HandleRegister).Methods("POST")
	v1.HandleFunc("/auth/login", app.HandleLogin).Methods("POST")
	v1.HandleFunc("/auth/refresh", app.HandleRefresh).Methods("POST")
	v1.HandleFunc("/auth/logout", app.HandleLogout).Methods("POST")
	v1.HandleFunc("/auth/validate", app.HandleTokenValidate).Methods("GET")
	v1.HandleFunc("/auth/introspect", app.HandleTokenIntrospect).Methods("POST")
	v1.HandleFunc("/auth/revoke", app.HandleRevokeToken).Methods("POST")

	// Admin endpoints (for managing applications)
	admin := v1.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/applications", app.HandleCreateApplication).Methods("POST")
	admin.HandleFunc("/applications", app.HandleGetApplications).Methods("GET")

	// Legacy endpoints (backward compatibility, will be deprecated)
	legacy := r.PathPrefix("/api/auth").Subrouter()
	legacy.Use(app.APIKeyAuth)
	legacy.Use(app.RateLimit)
	legacy.HandleFunc("/register", app.HandleRegister).Methods("POST")
	legacy.HandleFunc("/login", app.HandleLogin).Methods("POST")
	legacy.HandleFunc("/refresh", app.HandleRefresh).Methods("POST")
	legacy.HandleFunc("/logout", app.HandleLogout).Methods("POST")

	srv := &http.Server{Handler: r, Addr: ":" + c.Port, ReadTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second}

	go func() {
		fmt.Println("Starting Go server on", c.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if closer, ok := app.DB.(interface{ close() error }); ok {
		_ = closer.close()
	}
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown failed:%+v", err)
	}
	fmt.Println("Server exited properly")
}
