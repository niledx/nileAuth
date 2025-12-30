package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/example/nileauth/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	var (
		command = flag.String("command", "up", "Migration command: up, down, version, force")
		steps   = flag.Int("steps", 0, "Number of migration steps (for up/down)")
		version = flag.Uint("version", 0, "Target version (for force command)")
	)
	flag.Parse()

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	if cfg.DBAdapter != "postgres" {
		log.Fatalf("Migrations only work with PostgreSQL. Current adapter: %s", cfg.DBAdapter)
	}

	dsn, err := cfg.BuildPostgresDSN()
	if err != nil {
		log.Fatalf("PostgreSQL config error: %v", err)
	}

	migrationsDir := "./migrations"
	if len(os.Args) > 1 && os.Args[len(os.Args)-1] != "" {
		// Allow migrations dir as last argument
		if _, err := os.Stat(os.Args[len(os.Args)-1]); err == nil {
			migrationsDir = os.Args[len(os.Args)-1]
		}
	}

	switch *command {
	case "up":
		if err := runMigration(migrationsDir, dsn, true, *steps); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Println("✓ Migrations applied successfully")
	case "down":
		if err := runMigration(migrationsDir, dsn, false, *steps); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Println("✓ Migrations rolled back successfully")
	case "version":
		v, dirty, err := getMigrationVersion(migrationsDir, dsn)
		if err != nil {
			log.Fatalf("Failed to get version: %v", err)
		}
		if dirty {
			fmt.Printf("⚠ Database is in a dirty state (version %d)\n", v)
			os.Exit(1)
		}
		fmt.Printf("Current migration version: %d\n", v)
	case "force":
		if *version == 0 {
			log.Fatal("Version required for force command (use -version flag)")
		}
		if err := forceMigrationVersion(migrationsDir, dsn, int(*version)); err != nil {
			log.Fatalf("Force migration failed: %v", err)
		}
		fmt.Printf("✓ Forced database to version %d\n", *version)
	default:
		log.Fatalf("Unknown command: %s (supported: up, down, version, force)", *command)
	}
}

func runMigration(migrationsDir, dsn string, up bool, steps int) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("opening database connection: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("creating migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsDir, "postgres", driver)
	if err != nil {
		return fmt.Errorf("creating migrate instance: %w", err)
	}

	if steps > 0 {
		stepCount := steps
		if !up {
			stepCount = -steps
		}
		if err := m.Steps(stepCount); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("applying migrations: %w", err)
		}
	} else {
		if up {
			if err := m.Up(); err != nil && err != migrate.ErrNoChange {
				return fmt.Errorf("applying migrations: %w", err)
			}
		} else {
			if err := m.Down(); err != nil {
				return fmt.Errorf("rolling back migrations: %w", err)
			}
		}
	}
	return nil
}

func getMigrationVersion(migrationsDir, dsn string) (uint, bool, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return 0, false, fmt.Errorf("opening database connection: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("creating migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsDir, "postgres", driver)
	if err != nil {
		return 0, false, fmt.Errorf("creating migrate instance: %w", err)
	}

	version, dirty, err := m.Version()
	if err == migrate.ErrNilVersion {
		return 0, false, nil
	}
	return version, dirty, err
}

func forceMigrationVersion(migrationsDir, dsn string, version int) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("opening database connection: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("creating migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsDir, "postgres", driver)
	if err != nil {
		return fmt.Errorf("creating migrate instance: %w", err)
	}

	if err := m.Force(version); err != nil {
		return fmt.Errorf("forcing version: %w", err)
	}
	return nil
}

