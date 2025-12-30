# Database Migration Guide

## Quick Start

PostgreSQL is now the **default** database. Migrations run automatically on startup.

## Automatic Migrations

When you start the application, migrations are automatically applied:

```bash
go run .
```

You'll see output like:
```
Applying database migrations...
Migrated from version 0 to 2
Migrations applied successfully
Connected to PostgreSQL database
Starting Go server on 8080
```

## Manual Migration Commands

### Using Make (Recommended)

```bash
cd go

# Apply all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Check current version
make migrate-version

# Force to specific version (use with caution)
make migrate-force VERSION=1
```

### Using Migration Script

```bash
cd go
./migrate.sh up         # Apply all migrations
./migrate.sh down       # Rollback last migration
./migrate.sh version    # Show current version
./migrate.sh force 1    # Force to version 1
```

### Using Go Directly

```bash
# Apply migrations
go run ./go/cmd/migrate -command up

# Rollback
go run ./go/cmd/migrate -command down -steps 1

# Check version
go run ./go/cmd/migrate -command version
```

## Migration Files

All migrations are in `go/migrations/`:

- `V1__create_tables.up.sql` - Creates users and refresh_tokens tables
- `V1__create_tables.down.sql` - Rollback for V1
- `V2__add_enterprise_features.up.sql` - Adds applications, scopes, and enterprise features
- `V2__add_enterprise_features.down.sql` - Rollback for V2

## Configuration

Set PostgreSQL connection via environment variables:

**Option 1: DSN**
```bash
export POSTGRES_DSN="postgres://user:pass@localhost/nileauth?sslmode=disable"
```

**Option 2: Individual Parameters**
```bash
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=nile
export POSTGRES_PASSWORD=nilepass
export POSTGRES_DB=nileauth
export POSTGRES_SSLMODE=disable
```

## Docker Compose

The `docker-compose.yml` includes both PostgreSQL and the auth API:

```bash
# Start everything
docker-compose up -d

# View logs
docker-compose logs -f auth-api

# Stop
docker-compose down
```

Migrations run automatically when the API starts.

## Troubleshooting

### Migration fails on startup

If migrations fail, check:
1. PostgreSQL is running and accessible
2. Database exists
3. User has proper permissions
4. Connection string is correct

### Database in dirty state

If you see "database is in a dirty state":
1. Check what went wrong in the migration
2. Fix the issue manually if needed
3. Use `force` command to set version (use with caution)

### Rollback migrations

```bash
# Rollback last migration
make migrate-down

# Rollback multiple
go run ./go/cmd/migrate -command down -steps 2
```

## Best Practices

1. **Always backup** before running migrations in production
2. **Test migrations** in staging first
3. **Check version** before deploying: `make migrate-version`
4. **Monitor logs** during migration
5. **Never force** unless recovering from corruption

