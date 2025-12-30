# Nile Auth Service - Enterprise Authentication API

A centralized, enterprise-grade authentication service built in Go that serves as a single point of contact for authentication across all applications in your domain.

## Features

- **Multi-Application Support**: Register and manage multiple applications/clients
- **API Key Authentication**: Secure service-to-service authentication
- **Rate Limiting**: Per-application rate limiting to prevent abuse
- **CORS Support**: Configurable CORS per application
- **Token Management**: JWT access tokens and refresh tokens with rotation
- **Token Introspection**: OAuth 2.0 compliant token introspection
- **Token Validation**: Validate access tokens
- **Security Headers**: Built-in security headers (HSTS, XSS protection, etc.)
- **Structured Error Responses**: Consistent error format across all endpoints
- **Database Support**: PostgreSQL (default), SQLite, and in-memory storage
- **Easy Migrations**: Automatic migrations on startup + CLI tool for manual control
- **Health Checks**: `/health` and `/ready` endpoints for monitoring

## Quick Start

### Prerequisites

- Go 1.20 or later
- PostgreSQL 12+ (default, recommended for production)

### Configuration

PostgreSQL is now the **default** database. Set environment variables:

**Option 1: Using DSN (recommended)**
```bash
export PORT=8080
export JWT_SECRET=your-secret-key-here
export POSTGRES_DSN="postgres://user:pass@localhost/nileauth?sslmode=disable"
```

**Option 2: Using individual connection parameters**
```bash
export PORT=8080
export JWT_SECRET=your-secret-key-here
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=nile
export POSTGRES_PASSWORD=nilepass
export POSTGRES_DB=nileauth
export POSTGRES_SSLMODE=disable
```

**Alternative: SQLite (for development only)**
```bash
export DB_ADAPTER=sqlite
export SQLITE_FILE=./data/nile_go.db
```

### Run Locally

1. **Start PostgreSQL** (if not already running):
```bash
# Using docker-compose
docker-compose up -d db

# Or use your own PostgreSQL instance
```

2. **Run the application** (migrations run automatically):
```bash
cd go
go run .
```

The application will automatically apply database migrations on startup.

### Docker

**Using docker-compose (recommended):**
```bash
# Start both database and API
docker-compose up -d

# View logs
docker-compose logs -f auth-api
```

**Manual Docker build:**
```bash
# Build from repo root
docker build -t nile-auth -f go/Dockerfile .

# Run with PostgreSQL
docker run -e JWT_SECRET=your-secret \
  -e POSTGRES_HOST=host.docker.internal \
  -e POSTGRES_USER=nile \
  -e POSTGRES_PASSWORD=nilepass \
  -e POSTGRES_DB=nileauth \
  -p 8080:8080 nile-auth
```

## API Documentation

### Base URL

- Production: `https://auth.yourdomain.com`
- Development: `http://localhost:8080`

### Authentication

All API endpoints (except `/health` and `/ready`) require an API key:

**Header:**
```
X-API-Key: your-api-key-here
```

**OR Authorization Header:**
```
Authorization: Bearer your-api-key-here
```

### API Versioning

The API uses versioned endpoints:
- **v1**: `/api/v1/*` (current, recommended)
- **Legacy**: `/api/auth/*` (deprecated, maintained for backward compatibility)

---

## Endpoints

### Health & Status

#### GET `/health`
Health check endpoint (no authentication required)

**Response:**
```json
{
  "status": "ok"
}
```

#### GET `/ready`
Readiness check endpoint (no authentication required)

**Response:**
```json
{
  "ready": true
}
```

---

### Authentication Endpoints

#### POST `/api/v1/auth/register`

Register a new user.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Response (201):**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com"
  },
  "accessToken": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "abc123def456..."
}
```

**Errors:**
- `400 INVALID_REQUEST`: Missing email or password
- `409 USER_EXISTS`: User already exists

#### POST `/api/v1/auth/login`

Authenticate a user and receive tokens.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Response (200):**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com"
  },
  "accessToken": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "abc123def456..."
}
```

**Errors:**
- `400 INVALID_REQUEST`: Invalid request body
- `401 INVALID_CREDENTIALS`: Invalid email or password

#### POST `/api/v1/auth/refresh`

Refresh an access token using a refresh token.

**Request:**
```json
{
  "refreshToken": "abc123def456..."
}
```

**Response (200):**
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "new_refresh_token_here"
}
```

**Errors:**
- `400 INVALID_REQUEST`: Missing refresh token
- `401 INVALID_TOKEN`: Invalid or expired refresh token
- `401 TOKEN_REUSE_DETECTED`: Token reuse detected (security breach)

#### POST `/api/v1/auth/logout`

Revoke a refresh token.

**Request:**
```json
{
  "refreshToken": "abc123def456..."
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "revoked": true
  }
}
```

#### GET `/api/v1/auth/validate`

Validate an access token.

**Query Parameters:**
- `token` (optional): Token to validate (can also use `Authorization: Bearer <token>`)

**Response (200):**
```json
{
  "success": true,
  "data": {
    "valid": true,
    "userId": 1,
    "exp": 1234567890
  }
}
```

**Errors:**
- `400 INVALID_REQUEST`: Token not provided
- `401 INVALID_TOKEN`: Token is invalid or expired

#### POST `/api/v1/auth/introspect`

OAuth 2.0 token introspection endpoint.

**Request:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response (200):**
```json
{
  "active": true,
  "userId": 1,
  "expiresAt": 1234567890,
  "scopes": ["read:user"],
  "clientId": "app-123"
}
```

#### POST `/api/v1/auth/revoke`

Revoke a specific token.

**Request:**
```json
{
  "token": "abc123def456..."
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "revoked": true
  }
}
```

---

### Admin Endpoints

#### POST `/api/v1/admin/applications`

Register a new application/client.

**Request:**
```json
{
  "name": "My Application",
  "domain": "app.example.com",
  "rate_limit_per_minute": 100,
  "allowed_origins": ["https://app.example.com", "https://admin.example.com"]
}
```

**Response (201):**
```json
{
  "success": true,
  "data": {
    "application": {
      "id": 1,
      "name": "My Application",
      "domain": "app.example.com",
      "api_key_prefix": "a1b2c3d4",
      "rate_limit_per_minute": 100,
      "allowed_origins": ["https://app.example.com"]
    },
    "api_key": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6"
  }
}
```

**⚠️ Important:** The `api_key` is only returned once on creation. Store it securely!

#### GET `/api/v1/admin/applications`

List all applications (coming soon).

---

## Error Responses

All errors follow a consistent format:

```json
{
  "error_code": "ERROR_CODE",
  "error_message": "Human-readable error message"
}
```

**Common Error Codes:**
- `INVALID_REQUEST`: Bad request format
- `UNAUTHORIZED`: Missing or invalid API key
- `INVALID_CREDENTIALS`: Invalid email/password
- `INVALID_TOKEN`: Invalid or expired token
- `TOKEN_EXPIRED`: Token has expired
- `TOKEN_REUSE_DETECTED`: Security breach detected
- `USER_EXISTS`: User already registered
- `RATE_LIMIT_EXCEEDED`: Too many requests
- `INTERNAL_ERROR`: Server error

---

## Application Management

### Creating an Application

1. Call `POST /api/v1/admin/applications` to register your application
2. Store the returned `api_key` securely (it's only shown once!)
3. Use the `api_key` in all subsequent requests via `X-API-Key` header

### Rate Limiting

Each application has a configurable rate limit (default: 100 requests/minute). When exceeded, you'll receive:

```json
{
  "error_code": "RATE_LIMIT_EXCEEDED",
  "error_message": "Rate limit exceeded"
}
```

### CORS Configuration

Set `allowed_origins` when creating an application to enable CORS for specific domains. Use `["*"]` to allow all origins (not recommended for production).

---

## Database Migrations

### Automatic Migrations

Migrations are **automatically applied** when the application starts. The application will:
- Check the current migration version
- Apply any pending migrations
- Log migration status

### Manual Migration Management

For more control, use the migration CLI tool:

**Using Make (recommended):**
```bash
cd go
make migrate-up        # Apply all pending migrations
make migrate-down       # Rollback last migration
make migrate-version   # Show current version
make migrate-force VERSION=1  # Force to specific version
```

**Using the migration script:**
```bash
cd go
chmod +x migrate.sh
./migrate.sh up         # Apply all migrations
./migrate.sh down       # Rollback last migration
./migrate.sh version    # Show current version
./migrate.sh force 1    # Force to version 1
```

**Using Go directly:**
```bash
# Apply migrations
go run ./go/cmd/migrate -command up

# Rollback
go run ./go/cmd/migrate -command down

# Check version
go run ./go/cmd/migrate -command version

# Force version (use with caution)
go run ./go/cmd/migrate -command force -version 1
```

**Build migration tool:**
```bash
go build -o bin/migrate ./go/cmd/migrate
./bin/migrate -command up
```

### Migration Files

Migrations are located in `go/migrations/`:
- `V1__create_tables.up.sql` - Initial schema
- `V1__create_tables.down.sql` - Rollback for V1
- `V2__add_enterprise_features.up.sql` - Enterprise features
- `V2__add_enterprise_features.down.sql` - Rollback for V2

### Migration Best Practices

1. **Always backup** before running migrations in production
2. **Test migrations** in a staging environment first
3. **Check version** before deploying: `make migrate-version`
4. **Monitor logs** during migration for any errors
5. **Never force** migrations unless absolutely necessary (database corruption recovery)

### SQLite

SQLite migrations are automatically applied. The database file is created automatically if it doesn't exist. **Note:** SQLite doesn't support all enterprise features - use PostgreSQL for production.

---

## Security Best Practices

1. **JWT Secret**: Use a strong, randomly generated secret (minimum 32 characters)
2. **API Keys**: Store API keys securely (environment variables, secret managers)
3. **HTTPS**: Always use HTTPS in production
4. **Rate Limiting**: Configure appropriate rate limits per application
5. **CORS**: Restrict `allowed_origins` to specific domains
6. **Token Expiration**: Access tokens expire in 1 hour, refresh tokens in 30 days
7. **Token Rotation**: Refresh tokens are rotated on each use
8. **Token Reuse Detection**: Reusing a refresh token revokes all tokens for that user

---

## Integration Examples

### JavaScript/TypeScript

```javascript
const API_KEY = 'your-api-key-here';
const BASE_URL = 'https://auth.yourdomain.com';

// Register
const register = async (email, password) => {
  const response = await fetch(`${BASE_URL}/api/v1/auth/register`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': API_KEY
    },
    body: JSON.stringify({ email, password })
  });
  return response.json();
};

// Login
const login = async (email, password) => {
  const response = await fetch(`${BASE_URL}/api/v1/auth/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': API_KEY
    },
    body: JSON.stringify({ email, password })
  });
  return response.json();
};

// Validate token
const validateToken = async (token) => {
  const response = await fetch(`${BASE_URL}/api/v1/auth/validate?token=${token}`, {
    headers: {
      'X-API-Key': API_KEY
    }
  });
  return response.json();
};
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

func RegisterUser(email, password, apiKey string) error {
    data := map[string]string{
        "email":    email,
        "password": password,
    }
    jsonData, _ := json.Marshal(data)
    
    req, _ := http.NewRequest("POST", "https://auth.yourdomain.com/api/v1/auth/register", 
        bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", apiKey)
    
    client := &http.Client{}
    resp, err := client.Do(req)
    // Handle response...
    return err
}
```

---

## Monitoring

### Health Checks

Use `/health` and `/ready` endpoints for:
- Load balancer health checks
- Kubernetes liveness/readiness probes
- Monitoring systems

### Logging

The service logs all requests with:
- HTTP method and path
- Response status code
- Request duration
- Application identifier (API key prefix)

---

## Production Deployment

### Recommended Setup

1. **Database**: Use PostgreSQL for production
2. **Secrets**: Store `JWT_SECRET` and API keys in a secret manager
3. **HTTPS**: Use a reverse proxy (nginx, Traefik) with SSL/TLS
4. **Monitoring**: Set up logging and metrics collection
5. **Backup**: Regular database backups
6. **Scaling**: Use a load balancer with multiple instances

### Environment Variables

**Required:**
```bash
JWT_SECRET=<strong-random-secret>  # Required in production
```

**PostgreSQL (default):**
```bash
PORT=8080
# Option 1: DSN
POSTGRES_DSN=postgres://user:pass@host:5432/nileauth?sslmode=require

# Option 2: Individual parameters
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=nile
POSTGRES_PASSWORD=secure-password
POSTGRES_DB=nileauth
POSTGRES_SSLMODE=require  # Use 'disable' for local dev, 'require' for production
```

**Optional:**
```bash
DB_ADAPTER=postgres  # Default, can be 'sqlite' or 'memory'
LOG_LEVEL=info        # Default: info
```

**Note:** If `DB_ADAPTER` is not set, PostgreSQL is used by default. The application will fail to start if PostgreSQL connection parameters are missing.

---

## Support

For issues, questions, or contributions, please refer to the main project repository.
