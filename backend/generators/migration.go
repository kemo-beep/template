package generators

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

// MigrationInfo represents information about a migration
type MigrationInfo struct {
	Version     string    `json:"version"`
	Description string    `json:"description"`
	AppliedAt   time.Time `json:"applied_at"`
	TableName   string    `json:"table_name"`
	Operation   string    `json:"operation"` // CREATE, DROP, ALTER
}

// MigrationGenerator handles migration-based code generation
type MigrationGenerator struct {
	db *gorm.DB
}

// NewMigrationGenerator creates a new migration generator
func NewMigrationGenerator(db *gorm.DB) *MigrationGenerator {
	return &MigrationGenerator{db: db}
}

// GenerateAPIsFromMigration generates CRUD APIs based on migration files
func (mg *MigrationGenerator) GenerateAPIsFromMigration(migrationFile string) error {
	// Read migration file
	content, err := os.ReadFile(migrationFile)
	if err != nil {
		return err
	}

	// Parse migration to extract table information
	tableInfo, err := mg.parseMigration(string(content))
	if err != nil {
		return err
	}

	// Generate schema from table info
	schema := mg.tableInfoToSchema(tableInfo)

	// Generate APIs using schema generator
	sg := NewSchemaGenerator(mg.db)
	return sg.GenerateFromSchema(schema)
}

// CleanupAPIsFromMigration removes generated APIs when migration is dropped
func (mg *MigrationGenerator) CleanupAPIsFromMigration(migrationFile string) error {
	// Read migration file
	content, err := os.ReadFile(migrationFile)
	if err != nil {
		return err
	}

	// Parse migration to extract table information
	tableInfo, err := mg.parseMigration(string(content))
	if err != nil {
		return err
	}

	// Remove generated files
	modelName := strings.Title(tableInfo.TableName)

	filesToRemove := []string{
		fmt.Sprintf("models/%s.go", strings.ToLower(modelName)),
		fmt.Sprintf("controllers/%s.go", strings.ToLower(modelName)),
		fmt.Sprintf("routes/%s_routes.go", strings.ToLower(modelName)),
	}

	for _, file := range filesToRemove {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}

// parseMigration parses a migration file to extract table information
func (mg *MigrationGenerator) parseMigration(content string) (*TableInfo, error) {
	// Look for CREATE TABLE statements
	createTableRegex := regexp.MustCompile(`CREATE TABLE\s+(\w+)\s*\(([^)]+)\)`)
	matches := createTableRegex.FindStringSubmatch(content)

	if len(matches) < 3 {
		return nil, fmt.Errorf("no CREATE TABLE statement found")
	}

	tableName := matches[1]
	columnsStr := matches[2]

	// Parse columns
	columns := mg.parseColumns(columnsStr)

	return &TableInfo{
		TableName: tableName,
		Columns:   columns,
	}, nil
}

// TableInfo represents table structure information
type TableInfo struct {
	TableName string
	Columns   []ColumnInfo
}

// ColumnInfo represents column information
type ColumnInfo struct {
	Name       string
	Type       string
	Nullable   bool
	PrimaryKey bool
	Unique     bool
	Default    string
	Comment    string
}

// parseColumns parses column definitions from SQL
func (mg *MigrationGenerator) parseColumns(columnsStr string) []ColumnInfo {
	var columns []ColumnInfo

	// Split by comma, but be careful with nested parentheses
	lines := strings.Split(columnsStr, ",")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip constraints for now
		if strings.Contains(strings.ToUpper(line), "CONSTRAINT") ||
			strings.Contains(strings.ToUpper(line), "PRIMARY KEY") ||
			strings.Contains(strings.ToUpper(line), "FOREIGN KEY") {
			continue
		}

		column := mg.parseColumn(line)
		if column.Name != "" {
			columns = append(columns, column)
		}
	}

	return columns
}

// parseColumn parses a single column definition
func (mg *MigrationGenerator) parseColumn(line string) ColumnInfo {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return ColumnInfo{}
	}

	name := parts[0]
	columnType := parts[1]

	column := ColumnInfo{
		Name:     name,
		Type:     columnType,
		Nullable: true,
	}

	// Check for constraints
	for i, part := range parts {
		upperPart := strings.ToUpper(part)
		switch upperPart {
		case "NOT", "NULL":
			if i > 0 && strings.ToUpper(parts[i-1]) == "NOT" {
				column.Nullable = false
			}
		case "PRIMARY", "KEY":
			if i > 0 && strings.ToUpper(parts[i-1]) == "PRIMARY" {
				column.PrimaryKey = true
			}
		case "UNIQUE":
			column.Unique = true
		case "DEFAULT":
			if i+1 < len(parts) {
				column.Default = parts[i+1]
			}
		}
	}

	return column
}

// tableInfoToSchema converts table info to schema model
func (mg *MigrationGenerator) tableInfoToSchema(tableInfo *TableInfo) SchemaModel {
	var fields []SchemaField

	for _, col := range tableInfo.Columns {
		field := SchemaField{
			Name:     strings.Title(col.Name),
			Type:     col.Type,
			GoType:   mg.sqlTypeToGoType(col.Type),
			JSONTag:  fmt.Sprintf("json:\"%s\"", col.Name),
			GormTag:  mg.generateGormTag(col),
			Required: !col.Nullable,
			Unique:   col.Unique,
			Comment:  col.Comment,
		}

		if col.PrimaryKey {
			field.GormTag = "gorm:\"primaryKey\""
		}

		fields = append(fields, field)
	}

	return SchemaModel{
		Name:          strings.Title(tableInfo.TableName),
		Package:       "models",
		TableName:     tableInfo.TableName,
		Fields:        fields,
		HasTimestamps: mg.hasTimestamps(fields),
		HasSoftDelete: mg.hasSoftDelete(fields),
		Comment:       fmt.Sprintf("Auto-generated model for %s table", tableInfo.TableName),
	}
}

// sqlTypeToGoType converts SQL types to Go types
func (mg *MigrationGenerator) sqlTypeToGoType(sqlType string) string {
	sqlType = strings.ToUpper(sqlType)

	switch {
	case strings.Contains(sqlType, "INT"):
		return "int"
	case strings.Contains(sqlType, "BIGINT"):
		return "int64"
	case strings.Contains(sqlType, "VARCHAR"), strings.Contains(sqlType, "TEXT"), strings.Contains(sqlType, "CHAR"):
		return "string"
	case strings.Contains(sqlType, "BOOLEAN"), strings.Contains(sqlType, "BOOL"):
		return "bool"
	case strings.Contains(sqlType, "DECIMAL"), strings.Contains(sqlType, "NUMERIC"), strings.Contains(sqlType, "FLOAT"):
		return "float64"
	case strings.Contains(sqlType, "TIMESTAMP"), strings.Contains(sqlType, "DATETIME"):
		return "time.Time"
	case strings.Contains(sqlType, "DATE"):
		return "time.Time"
	case strings.Contains(sqlType, "JSON"):
		return "json.RawMessage"
	default:
		return "string"
	}
}

// generateGormTag generates GORM tags for a column
func (mg *MigrationGenerator) generateGormTag(col ColumnInfo) string {
	var tags []string

	if col.PrimaryKey {
		tags = append(tags, "primaryKey")
	}

	if !col.Nullable {
		tags = append(tags, "not null")
	}

	if col.Unique {
		tags = append(tags, "unique")
	}

	if col.Default != "" {
		tags = append(tags, fmt.Sprintf("default:%s", col.Default))
	}

	if col.Comment != "" {
		tags = append(tags, fmt.Sprintf("comment:%s", col.Comment))
	}

	if len(tags) == 0 {
		return ""
	}

	return fmt.Sprintf("gorm:\"%s\"", strings.Join(tags, ";"))
}

// hasTimestamps checks if the model has timestamp fields
func (mg *MigrationGenerator) hasTimestamps(fields []SchemaField) bool {
	hasCreatedAt := false
	hasUpdatedAt := false

	for _, field := range fields {
		if strings.ToLower(field.Name) == "createdat" || strings.ToLower(field.Name) == "created_at" {
			hasCreatedAt = true
		}
		if strings.ToLower(field.Name) == "updatedat" || strings.ToLower(field.Name) == "updated_at" {
			hasUpdatedAt = true
		}
	}

	return hasCreatedAt && hasUpdatedAt
}

// hasSoftDelete checks if the model has soft delete field
func (mg *MigrationGenerator) hasSoftDelete(fields []SchemaField) bool {
	for _, field := range fields {
		if strings.ToLower(field.Name) == "deletedat" || strings.ToLower(field.Name) == "deleted_at" {
			return true
		}
	}
	return false
}

// ListMigrations lists all applied migrations
func (mg *MigrationGenerator) ListMigrations() ([]MigrationInfo, error) {
	var migrations []MigrationInfo

	// Query migration table (assuming standard migration table structure)
	rows, err := mg.db.Raw("SELECT version, description, applied_at FROM schema_migrations ORDER BY applied_at DESC").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var migration MigrationInfo
		if err := rows.Scan(&migration.Version, &migration.Description, &migration.AppliedAt); err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
	}

	return migrations, nil
}

// GenerateAPIsForAllTables generates APIs for all existing tables
func (mg *MigrationGenerator) GenerateAPIsForAllTables() error {
	// Get all table names
	var tables []string
	if err := mg.db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tables).Error; err != nil {
		return err
	}

	// Generate APIs for each table
	for _, tableName := range tables {
		// Skip system tables
		if strings.HasPrefix(tableName, "pg_") || strings.HasPrefix(tableName, "schema_") {
			continue
		}

		// Get table structure
		tableInfo, err := mg.getTableInfo(tableName)
		if err != nil {
			continue // Skip tables we can't process
		}

		// Generate schema and APIs
		schema := mg.tableInfoToSchema(tableInfo)
		sg := NewSchemaGenerator(mg.db)
		if err := sg.GenerateFromSchema(schema); err != nil {
			fmt.Printf("Warning: Failed to generate APIs for table %s: %v\n", tableName, err)
		}
	}

	return nil
}

// getTableInfo gets table structure from database
func (mg *MigrationGenerator) getTableInfo(tableName string) (*TableInfo, error) {
	var columns []ColumnInfo

	query := `
		SELECT 
			column_name,
			data_type,
			is_nullable,
			column_default,
			character_maximum_length
		FROM information_schema.columns 
		WHERE table_name = ? AND table_schema = 'public'
		ORDER BY ordinal_position
	`

	rows, err := mg.db.Raw(query, tableName).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var col ColumnInfo
		var isNullable string
		var maxLength sql.NullInt64

		if err := rows.Scan(&col.Name, &col.Type, &isNullable, &col.Default, &maxLength); err != nil {
			return nil, err
		}

		col.Nullable = isNullable == "YES"

		// Check if it's a primary key
		if err := mg.db.Raw(`
			SELECT COUNT(*) 
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
			WHERE tc.table_name = ? AND tc.constraint_type = 'PRIMARY KEY' AND kcu.column_name = ?
		`, tableName, col.Name).Scan(&col.PrimaryKey).Error; err != nil {
			col.PrimaryKey = false
		}

		columns = append(columns, col)
	}

	return &TableInfo{
		TableName: tableName,
		Columns:   columns,
	}, nil
}
