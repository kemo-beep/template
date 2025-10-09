package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"mobile-backend/config"

	_ "github.com/lib/pq"
)

type Migration struct {
	Version     string
	Description string
	Filename    string
	Content     string
}

func main() {
	// Connect to database
	if err := config.ConnectDB(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	db := config.GetDB()
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database connection:", err)
	}

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(sqlDB); err != nil {
		log.Fatal("Failed to create migrations table:", err)
	}

	// Get migrations directory
	migrationsDir := "./migrations"
	if len(os.Args) > 1 {
		migrationsDir = os.Args[1]
	}

	// Load and run migrations
	migrations, err := loadMigrations(migrationsDir)
	if err != nil {
		log.Fatal("Failed to load migrations:", err)
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(sqlDB)
	if err != nil {
		log.Fatal("Failed to get applied migrations:", err)
	}

	// Run pending migrations
	pendingMigrations := getPendingMigrations(migrations, appliedMigrations)

	if len(pendingMigrations) == 0 {
		fmt.Println("✅ No pending migrations to run")
		return
	}

	fmt.Printf("🔄 Found %d pending migrations\n", len(pendingMigrations))

	for _, migration := range pendingMigrations {
		fmt.Printf("Running migration: %s - %s\n", migration.Version, migration.Description)

		if err := runMigration(sqlDB, migration); err != nil {
			log.Fatalf("Failed to run migration %s: %v", migration.Version, err)
		}

		fmt.Printf("✅ Migration %s completed successfully\n", migration.Version)
	}

	fmt.Println("🎉 All migrations completed successfully!")
}

func createMigrationsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		description VARCHAR(255) NOT NULL,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(query)
	return err
}

func loadMigrations(migrationsDir string) ([]Migration, error) {
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		// Parse version and description from filename
		parts := strings.Split(file.Name(), "_")
		if len(parts) < 2 {
			continue
		}

		version := parts[0]
		description := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".sql")

		// Read migration content
		content, err := ioutil.ReadFile(filepath.Join(migrationsDir, file.Name()))
		if err != nil {
			return nil, err
		}

		migrations = append(migrations, Migration{
			Version:     version,
			Description: description,
			Filename:    file.Name(),
			Content:     string(content),
		})
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	query := "SELECT version FROM schema_migrations"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, nil
}

func getPendingMigrations(migrations []Migration, applied map[string]bool) []Migration {
	var pending []Migration
	for _, migration := range migrations {
		if !applied[migration.Version] {
			pending = append(pending, migration)
		}
	}
	return pending
}

func runMigration(db *sql.DB, migration Migration) error {
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration
	if _, err := tx.Exec(migration.Content); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record migration as applied
	query := "INSERT INTO schema_migrations (version, description, applied_at) VALUES ($1, $2, $3)"
	_, err = tx.Exec(query, migration.Version, migration.Description, time.Now())
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	return tx.Commit()
}
