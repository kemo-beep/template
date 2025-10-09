package generators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

// AutoMigration handles automatic migration and API generation
type AutoMigration struct {
	db            *gorm.DB
	mg            *MigrationGenerator
	sg            *SchemaGenerator
	migrationsDir string
}

// NewAutoMigration creates a new auto migration instance
func NewAutoMigration(db *gorm.DB, migrationsDir string) *AutoMigration {
	return &AutoMigration{
		db:            db,
		mg:            NewMigrationGenerator(db),
		sg:            NewSchemaGenerator(db),
		migrationsDir: migrationsDir,
	}
}

// RunMigration runs a migration and generates APIs
func (am *AutoMigration) RunMigration(migrationFile string) error {
	// Check if migration file exists
	if _, err := os.Stat(migrationFile); os.IsNotExist(err) {
		return fmt.Errorf("migration file not found: %s", migrationFile)
	}

	// Read migration file
	content, err := os.ReadFile(migrationFile)
	if err != nil {
		return err
	}

	// Check if it's a CREATE TABLE migration
	if strings.Contains(strings.ToUpper(string(content)), "CREATE TABLE") {
		// Generate APIs for the new table
		if err := am.mg.GenerateAPIsFromMigration(migrationFile); err != nil {
			return fmt.Errorf("failed to generate APIs: %w", err)
		}

		fmt.Printf("✅ Generated APIs for migration: %s\n", filepath.Base(migrationFile))
	}

	// Check if it's a DROP TABLE migration
	if strings.Contains(strings.ToUpper(string(content)), "DROP TABLE") {
		// Clean up APIs for the dropped table
		if err := am.mg.CleanupAPIsFromMigration(migrationFile); err != nil {
			return fmt.Errorf("failed to cleanup APIs: %w", err)
		}

		fmt.Printf("🗑️  Cleaned up APIs for migration: %s\n", filepath.Base(migrationFile))
	}

	return nil
}

// RunAllMigrations runs all pending migrations
func (am *AutoMigration) RunAllMigrations() error {
	// Get all migration files
	migrationFiles, err := filepath.Glob(filepath.Join(am.migrationsDir, "*.sql"))
	if err != nil {
		return err
	}

	// Sort migration files by name (assuming timestamp prefix)
	// This is a simple sort - you might want to implement proper migration versioning
	for _, migrationFile := range migrationFiles {
		if err := am.RunMigration(migrationFile); err != nil {
			fmt.Printf("❌ Failed to process migration %s: %v\n", filepath.Base(migrationFile), err)
			continue
		}
	}

	return nil
}

// WatchMigrations watches for new migration files and processes them
func (am *AutoMigration) WatchMigrations() error {
	// This is a simplified version - in production you'd want to use fsnotify
	// or similar for real file watching

	fmt.Println("🔍 Watching for migration files...")

	// Check for new migrations every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := am.RunAllMigrations(); err != nil {
			fmt.Printf("❌ Error processing migrations: %v\n", err)
		}
	}

	return nil
}

// GenerateMigrationTemplate creates a migration template
func (am *AutoMigration) GenerateMigrationTemplate(tableName string, operation string) error {
	timestamp := time.Now().Format("20060102150405")
	fileName := fmt.Sprintf("%s_%s_%s.sql", timestamp, operation, tableName)
	filePath := filepath.Join(am.migrationsDir, fileName)

	var template string
	switch strings.ToUpper(operation) {
	case "CREATE":
		template = am.generateCreateTableTemplate(tableName)
	case "DROP":
		template = am.generateDropTableTemplate(tableName)
	case "ALTER":
		template = am.generateAlterTableTemplate(tableName)
	default:
		return fmt.Errorf("unsupported operation: %s", operation)
	}

	if err := os.WriteFile(filePath, []byte(template), 0644); err != nil {
		return err
	}

	fmt.Printf("📝 Created migration template: %s\n", fileName)
	return nil
}

// generateCreateTableTemplate generates a CREATE TABLE template
func (am *AutoMigration) generateCreateTableTemplate(tableName string) string {
	return fmt.Sprintf(`-- Migration: Create %s table
-- Generated: %s

CREATE TABLE %s (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Add your columns here
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Add indexes if needed
    INDEX idx_%s_name (name),
    INDEX idx_%s_is_active (is_active)
);

-- Add comments
COMMENT ON TABLE %s IS 'Auto-generated table for %s';
COMMENT ON COLUMN %s.name IS 'Name of the %s';
COMMENT ON COLUMN %s.description IS 'Description of the %s';
COMMENT ON COLUMN %s.is_active IS 'Whether the %s is active';
`,
		tableName, time.Now().Format("2006-01-02 15:04:05"),
		tableName, tableName, tableName, tableName, tableName, tableName, tableName, tableName, tableName, tableName, tableName)
}

// generateDropTableTemplate generates a DROP TABLE template
func (am *AutoMigration) generateDropTableTemplate(tableName string) string {
	return fmt.Sprintf(`-- Migration: Drop %s table
-- Generated: %s

DROP TABLE IF EXISTS %s CASCADE;
`,
		tableName, time.Now().Format("2006-01-02 15:04:05"), tableName)
}

// generateAlterTableTemplate generates an ALTER TABLE template
func (am *AutoMigration) generateAlterTableTemplate(tableName string) string {
	return fmt.Sprintf(`-- Migration: Alter %s table
-- Generated: %s

-- Add your ALTER statements here
-- Example:
-- ALTER TABLE %s ADD COLUMN new_field VARCHAR(255);
-- ALTER TABLE %s DROP COLUMN old_field;
-- ALTER TABLE %s ALTER COLUMN existing_field TYPE VARCHAR(500);
`,
		tableName, time.Now().Format("2006-01-02 15:04:05"), tableName, tableName, tableName)
}

// ListPendingMigrations lists pending migrations
func (am *AutoMigration) ListPendingMigrations() ([]string, error) {
	var pending []string

	// Get all migration files
	migrationFiles, err := filepath.Glob(filepath.Join(am.migrationsDir, "*.sql"))
	if err != nil {
		return nil, err
	}

	// Check which ones haven't been applied
	// This is a simplified check - in production you'd track this in a database
	for _, file := range migrationFiles {
		// For now, just return all files
		// In a real implementation, you'd check against a migrations table
		pending = append(pending, filepath.Base(file))
	}

	return pending, nil
}

// CreateMigrationFromSchema creates a migration file from a schema
func (am *AutoMigration) CreateMigrationFromSchema(schema SchemaModel) error {
	timestamp := time.Now().Format("20060102150405")
	fileName := fmt.Sprintf("%s_create_%s.sql", timestamp, strings.ToLower(schema.TableName))
	filePath := filepath.Join(am.migrationsDir, fileName)

	// Generate CREATE TABLE statement
	createSQL := am.generateCreateTableFromSchema(schema)

	if err := os.WriteFile(filePath, []byte(createSQL), 0644); err != nil {
		return err
	}

	fmt.Printf("📝 Created migration from schema: %s\n", fileName)
	return nil
}

// generateCreateTableFromSchema generates CREATE TABLE SQL from schema
func (am *AutoMigration) generateCreateTableFromSchema(schema SchemaModel) string {
	var sql strings.Builder

	sql.WriteString(fmt.Sprintf("-- Migration: Create %s table\n", schema.TableName))
	sql.WriteString(fmt.Sprintf("-- Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	sql.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", schema.TableName))

	// Add ID and timestamps if needed
	if schema.HasTimestamps {
		sql.WriteString("    id SERIAL PRIMARY KEY,\n")
		sql.WriteString("    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,\n")
		sql.WriteString("    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,\n")
	}

	if schema.HasSoftDelete {
		sql.WriteString("    deleted_at TIMESTAMP NULL,\n")
	}

	// Add custom fields
	for i, field := range schema.Fields {
		sql.WriteString(fmt.Sprintf("    %s %s", strings.ToLower(field.Name), field.Type))

		if field.Required {
			sql.WriteString(" NOT NULL")
		}

		if field.Unique {
			sql.WriteString(" UNIQUE")
		}

		if i < len(schema.Fields)-1 || schema.HasTimestamps || schema.HasSoftDelete {
			sql.WriteString(",")
		}

		sql.WriteString("\n")
	}

	sql.WriteString(");\n\n")

	// Add comments
	if schema.Comment != "" {
		sql.WriteString(fmt.Sprintf("COMMENT ON TABLE %s IS '%s';\n", schema.TableName, schema.Comment))
	}

	// Add column comments
	for _, field := range schema.Fields {
		if field.Comment != "" {
			sql.WriteString(fmt.Sprintf("COMMENT ON COLUMN %s.%s IS '%s';\n",
				schema.TableName, strings.ToLower(field.Name), field.Comment))
		}
	}

	return sql.String()
}

// SetupMigrationTable creates the migration tracking table
func (am *AutoMigration) SetupMigrationTable() error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		description TEXT,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	return am.db.Exec(createTableSQL).Error
}

// RecordMigration records a migration as applied
func (am *AutoMigration) RecordMigration(version, description string) error {
	return am.db.Exec(`
		INSERT INTO schema_migrations (version, description, applied_at) 
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (version) DO NOTHING
	`, version, description).Error
}
