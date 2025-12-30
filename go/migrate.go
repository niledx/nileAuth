package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// ApplyMigrations runs migrations from a local migrations directory against the provided Postgres DSN.
func ApplyMigrations(migrationsDir, dbURL string) error {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("opening database connection: %w", err)
	}
	defer db.Close()

	// Test connection
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
	
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("checking migration version: %w", err)
	}
	
	if dirty {
		return fmt.Errorf("database is in a dirty state (version %d). Manual intervention required", version)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Printf("Database is up to date (version %d)", version)
			return nil
		}
		return fmt.Errorf("applying migrations: %w", err)
	}
	
	newVersion, _, _ := m.Version()
	if newVersion != version {
		log.Printf("Migrated from version %d to %d", version, newVersion)
	}
	
	return nil
}

// GetMigrationVersion returns the current migration version
func GetMigrationVersion(migrationsDir, dbURL string) (uint, bool, error) {
	db, err := sql.Open("postgres", dbURL)
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
