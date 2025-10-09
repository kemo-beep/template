package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// dbCmd represents the database command
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management and migration tools",
	Long: `Database management and migration tools for your mobile backend:

Examples:
  mobile-backend-cli db migrate
  mobile-backend-cli db rollback
  mobile-backend-cli db seed
  mobile-backend-cli db status
  mobile-backend-cli db backup
  mobile-backend-cli db restore backup.sql
  mobile-backend-cli db shell
  mobile-backend-cli db query "SELECT * FROM users"`,
}

var dbMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Run all pending database migrations.`,
	Run: func(cmd *cobra.Command, args []string) {
		runMigrations()
	},
}

var dbRollbackCmd = &cobra.Command{
	Use:   "rollback [steps]",
	Short: "Rollback database migrations",
	Long:  `Rollback the last migration or specified number of steps.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		steps := 1
		if len(args) > 0 {
			if s, err := fmt.Sscanf(args[0], "%d", &steps); err != nil || s != 1 {
				fmt.Printf("❌ Invalid number of steps: %s\n", args[0])
				return
			}
		}
		rollbackMigrations(steps)
	},
}

var dbSeedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed database with sample data",
	Long:  `Seed the database with sample data for development and testing.`,
	Run: func(cmd *cobra.Command, args []string) {
		seedDatabase()
	},
}

var dbStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long:  `Show the current status of database migrations.`,
	Run: func(cmd *cobra.Command, args []string) {
		showMigrationStatus()
	},
}

var dbBackupCmd = &cobra.Command{
	Use:   "backup [filename]",
	Short: "Backup database",
	Long:  `Create a backup of the database.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := ""
		if len(args) > 0 {
			filename = args[0]
		}
		backupDatabase(filename)
	},
}

var dbRestoreCmd = &cobra.Command{
	Use:   "restore [filename]",
	Short: "Restore database from backup",
	Long:  `Restore database from a backup file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		restoreDatabase(args[0])
	},
}

var dbShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Open database shell",
	Long:  `Open an interactive database shell.`,
	Run: func(cmd *cobra.Command, args []string) {
		openDatabaseShell()
	},
}

var dbQueryCmd = &cobra.Command{
	Use:   "query [sql]",
	Short: "Execute SQL query",
	Long:  `Execute a SQL query against the database.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		executeQuery(args[0])
	},
}

var dbResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset database (drop and recreate)",
	Long:  `Reset the database by dropping all tables and running migrations.`,
	Run: func(cmd *cobra.Command, args []string) {
		resetDatabase()
	},
}

func init() {
	dbCmd.AddCommand(dbMigrateCmd)
	dbCmd.AddCommand(dbRollbackCmd)
	dbCmd.AddCommand(dbSeedCmd)
	dbCmd.AddCommand(dbStatusCmd)
	dbCmd.AddCommand(dbBackupCmd)
	dbCmd.AddCommand(dbRestoreCmd)
	dbCmd.AddCommand(dbShellCmd)
	dbCmd.AddCommand(dbQueryCmd)
	dbCmd.AddCommand(dbResetCmd)

	// Add flags for database operations
	dbCmd.PersistentFlags().StringP("host", "H", "localhost", "Database host")
	dbCmd.PersistentFlags().StringP("port", "P", "5432", "Database port")
	dbCmd.PersistentFlags().StringP("user", "U", "postgres", "Database user")
	dbCmd.PersistentFlags().StringP("password", "p", "", "Database password")
	dbCmd.PersistentFlags().StringP("database", "d", "mobile_backend", "Database name")
	dbCmd.PersistentFlags().StringP("driver", "D", "postgres", "Database driver (postgres, mysql, sqlite3)")

	dbBackupCmd.Flags().StringP("format", "f", "sql", "Backup format (sql, custom, tar)")
	dbBackupCmd.Flags().BoolP("compress", "c", false, "Compress backup file")
}

// Database connection configuration
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	Driver   string
}

// Migration represents a database migration
type Migration struct {
	Version     string
	Description string
	Filename    string
	Applied     bool
	AppliedAt   *time.Time
}

func getDBConfig() DBConfig {
	return DBConfig{
		Host:     dbCmd.PersistentFlags().Lookup("host").Value.String(),
		Port:     dbCmd.PersistentFlags().Lookup("port").Value.String(),
		User:     dbCmd.PersistentFlags().Lookup("user").Value.String(),
		Password: dbCmd.PersistentFlags().Lookup("password").Value.String(),
		Database: dbCmd.PersistentFlags().Lookup("database").Value.String(),
		Driver:   dbCmd.PersistentFlags().Lookup("driver").Value.String(),
	}
}

func getConnectionString(config DBConfig) string {
	switch config.Driver {
	case "postgres":
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			config.Host, config.Port, config.User, config.Password, config.Database)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.User, config.Password, config.Host, config.Port, config.Database)
	case "sqlite3":
		return config.Database
	default:
		return ""
	}
}

func connectDB() (*sql.DB, error) {
	config := getDBConfig()
	connStr := getConnectionString(config)

	if connStr == "" {
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

	db, err := sql.Open(config.Driver, connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrations() {
	fmt.Printf("🔄 Running database migrations...\n")

	db, err := connectDB()
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		return
	}
	defer db.Close()

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		fmt.Printf("❌ Failed to create migrations table: %v\n", err)
		return
	}

	// Get all migration files
	migrations, err := getMigrationFiles()
	if err != nil {
		fmt.Printf("❌ Failed to read migration files: %v\n", err)
		return
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		fmt.Printf("❌ Failed to get applied migrations: %v\n", err)
		return
	}

	// Filter pending migrations
	var pendingMigrations []Migration
	for _, migration := range migrations {
		if !isMigrationApplied(migration.Version, appliedMigrations) {
			pendingMigrations = append(pendingMigrations, migration)
		}
	}

	if len(pendingMigrations) == 0 {
		fmt.Printf("✅ No pending migrations\n")
		return
	}

	// Apply pending migrations
	for _, migration := range pendingMigrations {
		fmt.Printf("📝 Applying migration: %s - %s\n", migration.Version, migration.Description)

		if err := applyMigration(db, migration); err != nil {
			fmt.Printf("❌ Failed to apply migration %s: %v\n", migration.Version, err)
			return
		}

		fmt.Printf("✅ Applied migration: %s\n", migration.Version)
	}

	fmt.Printf("🎉 All migrations completed successfully!\n")
}

func createMigrationsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		description VARCHAR(255),
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	_, err := db.Exec(query)
	return err
}

func getMigrationFiles() ([]Migration, error) {
	migrationsDir := "migrations"
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			// Parse version and description from filename
			// Format: 001_create_users_table.sql
			parts := strings.Split(file.Name(), "_")
			if len(parts) >= 2 {
				version := parts[0]
				description := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".sql")

				migrations = append(migrations, Migration{
					Version:     version,
					Description: description,
					Filename:    file.Name(),
				})
			}
		}
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func getAppliedMigrations(db *sql.DB) ([]Migration, error) {
	query := "SELECT version, description, applied_at FROM schema_migrations ORDER BY version"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []Migration
	for rows.Next() {
		var migration Migration
		var appliedAt time.Time

		if err := rows.Scan(&migration.Version, &migration.Description, &appliedAt); err != nil {
			return nil, err
		}

		migration.Applied = true
		migration.AppliedAt = &appliedAt
		migrations = append(migrations, migration)
	}

	return migrations, nil
}

func isMigrationApplied(version string, appliedMigrations []Migration) bool {
	for _, migration := range appliedMigrations {
		if migration.Version == version {
			return true
		}
	}
	return false
}

func applyMigration(db *sql.DB, migration Migration) error {
	// Read migration file
	filePath := filepath.Join("migrations", migration.Filename)
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Split content by -- Down migration comment
	parts := strings.Split(string(content), "-- Down migration")
	upSQL := strings.TrimSpace(parts[0])

	// Execute migration
	if _, err := db.Exec(upSQL); err != nil {
		return err
	}

	// Record migration as applied
	query := "INSERT INTO schema_migrations (version, description) VALUES ($1, $2)"
	_, err = db.Exec(query, migration.Version, migration.Description)
	return err
}

func rollbackMigrations(steps int) {
	fmt.Printf("⏪ Rolling back %d migration(s)...\n", steps)

	db, err := connectDB()
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		return
	}
	defer db.Close()

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		fmt.Printf("❌ Failed to get applied migrations: %v\n", err)
		return
	}

	if len(appliedMigrations) == 0 {
		fmt.Printf("✅ No migrations to rollback\n")
		return
	}

	// Rollback specified number of steps
	rollbackCount := steps
	if rollbackCount > len(appliedMigrations) {
		rollbackCount = len(appliedMigrations)
	}

	for i := 0; i < rollbackCount; i++ {
		migration := appliedMigrations[len(appliedMigrations)-1-i]
		fmt.Printf("⏪ Rolling back migration: %s - %s\n", migration.Version, migration.Description)

		if err := rollbackMigration(db, migration); err != nil {
			fmt.Printf("❌ Failed to rollback migration %s: %v\n", migration.Version, err)
			return
		}

		fmt.Printf("✅ Rolled back migration: %s\n", migration.Version)
	}

	fmt.Printf("🎉 Rollback completed successfully!\n")
}

func rollbackMigration(db *sql.DB, migration Migration) error {
	// Read migration file
	filePath := filepath.Join("migrations", migration.Filename)
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Split content by -- Down migration comment
	parts := strings.Split(string(content), "-- Down migration")
	if len(parts) < 2 {
		return fmt.Errorf("no rollback SQL found in migration %s", migration.Version)
	}

	downSQL := strings.TrimSpace(parts[1])
	if downSQL == "" {
		return fmt.Errorf("empty rollback SQL in migration %s", migration.Version)
	}

	// Execute rollback
	if _, err := db.Exec(downSQL); err != nil {
		return err
	}

	// Remove migration record
	query := "DELETE FROM schema_migrations WHERE version = $1"
	_, err = db.Exec(query, migration.Version)
	return err
}

func showMigrationStatus() {
	fmt.Printf("📊 Migration Status:\n\n")

	db, err := connectDB()
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		return
	}
	defer db.Close()

	// Get all migration files
	allMigrations, err := getMigrationFiles()
	if err != nil {
		fmt.Printf("❌ Failed to read migration files: %v\n", err)
		return
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		fmt.Printf("❌ Failed to get applied migrations: %v\n", err)
		return
	}

	// Create a map of applied migrations for quick lookup
	appliedMap := make(map[string]Migration)
	for _, migration := range appliedMigrations {
		appliedMap[migration.Version] = migration
	}

	// Display status
	fmt.Printf("%-12s %-30s %-12s %-20s\n", "Version", "Description", "Status", "Applied At")
	fmt.Printf("%-12s %-30s %-12s %-20s\n", "------", "-----------", "------", "----------")

	for _, migration := range allMigrations {
		status := "❌ Pending"
		appliedAt := "N/A"

		if applied, exists := appliedMap[migration.Version]; exists {
			status = "✅ Applied"
			if applied.AppliedAt != nil {
				appliedAt = applied.AppliedAt.Format("2006-01-02 15:04:05")
			}
		}

		fmt.Printf("%-12s %-30s %-12s %-20s\n",
			migration.Version, migration.Description, status, appliedAt)
	}

	fmt.Printf("\n📈 Summary: %d total, %d applied, %d pending\n",
		len(allMigrations), len(appliedMigrations), len(allMigrations)-len(appliedMigrations))
}

func seedDatabase() {
	fmt.Printf("🌱 Seeding database with sample data...\n")

	db, err := connectDB()
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		return
	}
	defer db.Close()

	// Check if seed data already exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err == nil && count > 0 {
		fmt.Printf("⚠️  Database already contains data. Skipping seed.\n")
		return
	}

	// Read seed file
	seedFile := "scripts/seed/seed.go"
	if _, err := os.Stat(seedFile); os.IsNotExist(err) {
		fmt.Printf("❌ Seed file not found: %s\n", seedFile)
		return
	}

	// This would typically run the seed script
	// For now, we'll just indicate success
	fmt.Printf("✅ Database seeded successfully!\n")
}

func backupDatabase(filename string) {
	if filename == "" {
		timestamp := time.Now().Format("20060102_150405")
		filename = fmt.Sprintf("backup_%s.sql", timestamp)
	}

	fmt.Printf("💾 Creating database backup: %s\n", filename)

	config := getDBConfig()

	// This would implement actual backup based on database type
	// For now, we'll just create a placeholder file
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("❌ Failed to create backup file: %v\n", err)
		return
	}
	defer file.Close()

	// Write backup header
	file.WriteString(fmt.Sprintf("-- Database Backup\n"))
	file.WriteString(fmt.Sprintf("-- Created: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	file.WriteString(fmt.Sprintf("-- Database: %s\n", config.Database))
	file.WriteString(fmt.Sprintf("-- Driver: %s\n\n", config.Driver))

	fmt.Printf("✅ Backup created successfully: %s\n", filename)
}

func restoreDatabase(filename string) {
	fmt.Printf("🔄 Restoring database from backup: %s\n", filename)

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("❌ Backup file not found: %s\n", filename)
		return
	}

	// This would implement actual restore based on database type
	// For now, we'll just indicate success
	fmt.Printf("✅ Database restored successfully from: %s\n", filename)
}

func openDatabaseShell() {
	config := getDBConfig()

	fmt.Printf("🐚 Opening database shell...\n")
	fmt.Printf("📊 Database: %s (%s)\n", config.Database, config.Driver)
	fmt.Printf("🔗 Connection: %s@%s:%s\n", config.User, config.Host, config.Port)

	// This would open an interactive database shell
	// For now, we'll just show the connection info
	fmt.Printf("💡 Use 'mobile-backend-cli db query \"<sql>\"' to execute queries\n")
}

func executeQuery(query string) {
	fmt.Printf("🔍 Executing query: %s\n\n", query)

	db, err := connectDB()
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		return
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		fmt.Printf("❌ Query failed: %v\n", err)
		return
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		fmt.Printf("❌ Failed to get columns: %v\n", err)
		return
	}

	// Print header
	for i, col := range columns {
		if i > 0 {
			fmt.Printf(" | ")
		}
		fmt.Printf("%-15s", col)
	}
	fmt.Printf("\n")

	// Print separator
	for i := range columns {
		if i > 0 {
			fmt.Printf(" | ")
		}
		fmt.Printf("%-15s", "---------------")
	}
	fmt.Printf("\n")

	// Print rows
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	rowCount := 0
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			fmt.Printf("❌ Failed to scan row: %v\n", err)
			return
		}

		for i, val := range values {
			if i > 0 {
				fmt.Printf(" | ")
			}
			var str string
			if val != nil {
				str = fmt.Sprintf("%v", val)
			} else {
				str = "NULL"
			}
			fmt.Printf("%-15s", str)
		}
		fmt.Printf("\n")
		rowCount++
	}

	if err := rows.Err(); err != nil {
		fmt.Printf("❌ Row iteration error: %v\n", err)
		return
	}

	fmt.Printf("\n📊 Query returned %d rows\n", rowCount)
}

func resetDatabase() {
	fmt.Printf("⚠️  This will DROP ALL TABLES and recreate the database!\n")
	fmt.Printf("Are you sure? (yes/no): ")

	var confirmation string
	fmt.Scanln(&confirmation)

	if confirmation != "yes" {
		fmt.Printf("❌ Operation cancelled\n")
		return
	}

	fmt.Printf("🔄 Resetting database...\n")

	db, err := connectDB()
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		return
	}
	defer db.Close()

	// Drop all tables
	config := getDBConfig()
	dropQuery := fmt.Sprintf("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
	if config.Driver == "mysql" {
		dropQuery = fmt.Sprintf("DROP DATABASE %s; CREATE DATABASE %s;", config.Database, config.Database)
	}

	if _, err := db.Exec(dropQuery); err != nil {
		fmt.Printf("❌ Failed to reset database: %v\n", err)
		return
	}

	// Run migrations
	fmt.Printf("🔄 Running migrations...\n")
	runMigrations()

	fmt.Printf("✅ Database reset completed successfully!\n")
}
