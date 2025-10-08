package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"mobile-backend/config"
	"mobile-backend/generators"
)

func main() {
	var (
		action        = flag.String("action", "", "Action to perform: generate, migrate, cleanup, template")
		schemaFile    = flag.String("schema", "", "Schema JSON file path")
		migrationFile = flag.String("migration", "", "Migration SQL file path")
		modelName     = flag.String("model", "", "Model name for cleanup")
		tableName     = flag.String("table", "", "Table name for template generation")
		operation     = flag.String("op", "create", "Operation: create, drop, alter")
		migrationsDir = flag.String("migrations", "./migrations", "Migrations directory")
	)
	flag.Parse()

	// Connect to database
	if err := config.ConnectDB(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	db := config.GetDB()
	sg := generators.NewSchemaGenerator(db)
	mg := generators.NewMigrationGenerator(db)
	am := generators.NewAutoMigration(db, *migrationsDir)

	switch *action {
	case "generate":
		if *schemaFile == "" {
			log.Fatal("Schema file is required for generation")
		}
		if err := sg.GenerateFromJSON(*schemaFile); err != nil {
			log.Fatal("Failed to generate from schema:", err)
		}
		fmt.Println("✅ APIs generated successfully!")

	case "migrate":
		if *migrationFile == "" {
			log.Fatal("Migration file is required")
		}
		if err := am.RunMigration(*migrationFile); err != nil {
			log.Fatal("Failed to run migration:", err)
		}
		fmt.Println("✅ Migration completed successfully!")

	case "migrate-all":
		if err := am.RunAllMigrations(); err != nil {
			log.Fatal("Failed to run all migrations:", err)
		}
		fmt.Println("✅ All migrations completed successfully!")

	case "generate-all":
		if err := mg.GenerateAPIsForAllTables(); err != nil {
			log.Fatal("Failed to generate APIs for all tables:", err)
		}
		fmt.Println("✅ APIs generated for all tables successfully!")

	case "cleanup":
		if *modelName == "" {
			log.Fatal("Model name is required for cleanup")
		}
		if err := mg.CleanupAPIsFromMigration(*modelName); err != nil {
			log.Fatal("Failed to cleanup model:", err)
		}
		fmt.Println("✅ Model cleanup completed successfully!")

	case "template":
		if *tableName == "" {
			log.Fatal("Table name is required for template generation")
		}
		if err := am.GenerateMigrationTemplate(*tableName, *operation); err != nil {
			log.Fatal("Failed to generate template:", err)
		}
		fmt.Println("✅ Migration template generated successfully!")

	case "watch":
		fmt.Println("🔍 Watching for migration files...")
		if err := am.WatchMigrations(); err != nil {
			log.Fatal("Failed to watch migrations:", err)
		}

	case "setup":
		// Create migrations directory
		if err := os.MkdirAll(*migrationsDir, 0755); err != nil {
			log.Fatal("Failed to create migrations directory:", err)
		}

		// Setup migration table
		if err := am.SetupMigrationTable(); err != nil {
			log.Fatal("Failed to setup migration table:", err)
		}

		fmt.Println("✅ Setup completed successfully!")

	default:
		fmt.Println("Mobile Backend API Generator")
		fmt.Println("Usage: go run cmd/generator/main.go -action <action> [options]")
		fmt.Println("")
		fmt.Println("Actions:")
		fmt.Println("  generate     Generate APIs from schema JSON file")
		fmt.Println("  migrate      Run a specific migration file")
		fmt.Println("  migrate-all  Run all pending migrations")
		fmt.Println("  generate-all Generate APIs for all existing tables")
		fmt.Println("  cleanup      Cleanup generated files for a model")
		fmt.Println("  template     Generate migration template")
		fmt.Println("  watch        Watch for new migration files")
		fmt.Println("  setup        Setup migrations directory and table")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  go run cmd/generator/main.go -action generate -schema product.json")
		fmt.Println("  go run cmd/generator/main.go -action migrate -migration 001_create_products.sql")
		fmt.Println("  go run cmd/generator/main.go -action generate-all")
		fmt.Println("  go run cmd/generator/main.go -action cleanup -model product")
		fmt.Println("  go run cmd/generator/main.go -action template -table products -op create")
	}
}
