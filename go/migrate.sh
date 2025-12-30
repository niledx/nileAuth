#!/bin/bash
# Migration helper script for Nile Auth Service

set -e

COMMAND=${1:-up}
STEPS=${2:-0}
VERSION=${3:-0}

# Load environment variables from .env if it exists
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Default PostgreSQL connection if not set
export POSTGRES_HOST=${POSTGRES_HOST:-localhost}
export POSTGRES_PORT=${POSTGRES_PORT:-5432}
export POSTGRES_USER=${POSTGRES_USER:-nile}
export POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-nilepass}
export POSTGRES_DB=${POSTGRES_DB:-nileauth}
export POSTGRES_SSLMODE=${POSTGRES_SSLMODE:-disable}

case "$COMMAND" in
    up)
        echo "Applying migrations..."
        go run ./cmd/migrate -command up -steps $STEPS
        ;;
    down)
        echo "Rolling back migrations..."
        go run ./cmd/migrate -command down -steps $STEPS
        ;;
    version)
        echo "Current migration version:"
        go run ./cmd/migrate -command version
        ;;
    force)
        if [ "$VERSION" = "0" ]; then
            echo "Error: Version required for force command"
            echo "Usage: ./migrate.sh force <version>"
            exit 1
        fi
        echo "Forcing migration version to $VERSION..."
        go run ./cmd/migrate -command force -version $VERSION
        ;;
    *)
        echo "Usage: $0 {up|down|version|force} [steps] [version]"
        echo ""
        echo "Commands:"
        echo "  up       - Apply all pending migrations"
        echo "  down     - Rollback last migration"
        echo "  version  - Show current migration version"
        echo "  force    - Force migration version (requires version number)"
        echo ""
        echo "Examples:"
        echo "  $0 up              # Apply all migrations"
        echo "  $0 up 2            # Apply 2 migrations"
        echo "  $0 down            # Rollback last migration"
        echo "  $0 down 2          # Rollback 2 migrations"
        echo "  $0 version         # Show current version"
        echo "  $0 force 1         # Force version to 1"
        exit 1
        ;;
esac

