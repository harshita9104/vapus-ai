package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"

	// Import the existing migrations
	"github.com/vapusdata-ecosystem/vapusai/core/app/migration/migrations"
)

func main() {
	// Load .env file if it exists
	loadEnvFile()

	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := os.Args[1]

	switch command {
	case "init":
		runInit()
	case "create", "creation":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run ./simple-main.go creation <migration_name>")
			return
		}
		createGoMigrationFiles(strings.Join(os.Args[2:], "_"))
	case "run", ":run":
		runAllMigrations()
	case "revert", "rollback":
		revertLastMigration()
	case "status":
		showMigrationStatus()
	case "verify":
		runVerify()
	case "unlock":
		runUnlock()
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
	}
}

func printHelp() {
	fmt.Println(`VapusAI Go-Based Migration Tool

Usage:
  go run ./simple-main.go <command>

Commands:
  init         Initialize migration table  
  creation     Create new Go migration files (up & down)
  :run         Run all pending migrations (top to bottom)
  revert       Rollback last migration (runs down file)
  status       Show migration status and sequence
  verify       Verify database changes
  unlock       Unlock migrations when stuck
  help         Show this help

Environment Variables:
  DB_HOST     Database host
  DB_PORT     Database port  
  DB_USER     Database username
  DB_PASSWORD Database password (required)
  DB_NAME     Database name
  DB_SSLMODE  SSL mode

Examples:
  go run ./simple-main.go creation add_user_table
  go run ./simple-main.go :run
  go run ./simple-main.go revert
  go run ./simple-main.go status`)
}

func getDBConnection() (*bun.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "vapusaidev")
	sslmode := getEnv("DB_SSLMODE", "disable")

	if password == "" {
		return nil, fmt.Errorf("DB_PASSWORD environment variable is required")
	}

	// URL encode the password to handle special characters
	encodedPassword := url.QueryEscape(password)
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, encodedPassword, host, port, dbname, sslmode)

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	fmt.Printf("✅ Connected to database: %s@%s:%s/%s\n", user, host, port, dbname)
	return db, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func runInit() {
	fmt.Println("Initializing migration table...")

	db, err := getDBConnection()
	if err != nil {
		log.Fatal("❌ ", err)
	}
	defer db.Close()

	migrator := migrate.NewMigrator(db, migrations.Migrations)

	if err := migrator.Init(context.Background()); err != nil {
		log.Fatal("❌ Failed to initialize migrations: ", err)
	}

	fmt.Println("✅ Migration table initialized successfully!")
}

// createGoMigrationFiles creates both up and down Go migration files with timestamp
func createGoMigrationFiles(name string) {
	fmt.Printf("Creating Go migration files: %s\n", name)

	// Generate timestamp for unique filename
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	fileName := fmt.Sprintf("%s_%s.go", timestamp, name)

	migrationsDir := "/home/harshita/workspace/vapus-ai/internals/app/migration/migrations"

	// Create migrations directory if it doesn't exist
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create migrations directory: %v", err)
	}

	filePath := filepath.Join(migrationsDir, fileName)

	// Generate table name and model name from migration name
	// Convert add_products_table -> AddProductsTable
	words := strings.Split(name, "_")
	var modelNameParts []string
	for _, word := range words {
		modelNameParts = append(modelNameParts, capitalizeFirst(word))
	}
	modelName := strings.Join(modelNameParts, "") + "Model"

	// Bun creates table names by converting ModelName to snake_case and pluralizing
	// TestFinalMigrationModel -> test_final_migration_models
	// So for test_final_migration, model is TestFinalMigrationModel, table is test_final_migration_models
	tableName := name + "_models"

	// Template for Go migration (both up and down in one file)
	template := `package migrations

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Printf(" [UP] Running migration: ` + name + `\n")

		// Check if table exists before creating
		var exists bool
		err := db.NewSelect().
			ColumnExpr("true").
			TableExpr("information_schema.tables").
			Where("table_name = ? AND table_schema = 'public'", "` + tableName + `").
			Scan(ctx, &exists)

		if err != nil && err.Error() != "sql: no rows in result set" {
			return fmt.Errorf("failed to check table existence: %w", err)
		}

		if !exists {
			// Create ` + tableName + ` table
			_, err = db.NewCreateTable().
				Model((*` + modelName + `)(nil)).
				IfNotExists().
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to create table: %w", err)
			}

			// Insert sample data
			sampleData := &` + modelName + `{
				Name:        "Sample ` + name + `",
				Description: "Created by Go migration",
				CreatedAt:   time.Now(),
			}
			_, err = db.NewInsert().Model(sampleData).Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to insert sample data: %w", err)
			}

			fmt.Printf("   ✅ Migration completed successfully - created ` + tableName + ` table\n")
		} else {
			fmt.Printf("   ⚠️  Table ` + tableName + ` already exists, skipping\n")
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Printf("  [DOWN] Rolling back migration: ` + name + `\n")

		// Check if table exists before dropping
		var exists bool
		err := db.NewSelect().
			ColumnExpr("true").
			TableExpr("information_schema.tables").
			Where("table_name = ? AND table_schema = 'public'", "` + tableName + `").
			Scan(ctx, &exists)

		if err != nil && err.Error() != "sql: no rows in result set" {
			return fmt.Errorf("failed to check table existence: %w", err)
		}

		if exists {
			// Drop the ` + tableName + ` table
			_, err = db.NewDropTable().
				Model((*` + modelName + `)(nil)).
				IfExists().
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to drop table: %w", err)
			}
			fmt.Printf("   ✅ Rollback completed successfully - dropped ` + tableName + ` table\n")
		} else {
			fmt.Printf("   ⚠️  Table ` + tableName + ` doesn't exist, skipping\n")
		}

		return nil
	})
}

// ` + modelName + ` - Model for ` + name + ` migration
type ` + modelName + ` struct {
	ID          int64     ` + "`bun:\"id,pk,autoincrement\"`" + `
	Name        string    ` + "`bun:\"name,notnull\"`" + `
	Description string    ` + "`bun:\"description\"`" + `
	CreatedAt   time.Time ` + "`bun:\"created_at,notnull,default:current_timestamp\"`" + `
}
`

	// Write migration file
	if err := os.WriteFile(filePath, []byte(template), 0644); err != nil {
		log.Fatalf("❌ Failed to create migration file: %v", err)
	}

	fmt.Printf("✅ Created migration file: %s\n", filePath)
	fmt.Println("� Migration is ready to run! Use './simple-main :run' to apply it")
}

// runAllMigrations runs all pending migrations from top to bottom
func runAllMigrations() {
	fmt.Println(" Running all pending migrations (top to bottom)...")

	db, err := getDBConnection()
	if err != nil {
		log.Fatal("❌ ", err)
	}
	defer db.Close()

	migrator := migrate.NewMigrator(db, migrations.Migrations)

	// Check if migration table exists, create if not
	if err := migrator.Init(context.Background()); err != nil {
		log.Fatalf("❌ Failed to initialize migration table: %v", err)
	}

	// Lock migrations to prevent concurrent runs
	if err := migrator.Lock(context.Background()); err != nil {
		log.Fatalf("❌ Failed to lock migrations: %v", err)
	}
	defer migrator.Unlock(context.Background())

	// Get migration status before running
	ms, err := migrator.MigrationsWithStatus(context.Background())
	if err != nil {
		log.Fatalf("❌ Failed to get migration status: %v", err)
	}

	unapplied := ms.Unapplied()
	if len(unapplied) == 0 {
		fmt.Println("✅ No pending migrations to run (database is up to date)")
		return
	}

	fmt.Printf(" Found %d pending migration(s)\n", len(unapplied))

	// Run migrations
	group, err := migrator.Migrate(context.Background())
	if err != nil {
		log.Fatalf("❌ Failed to run migrations: %v", err)
	}

	if !group.IsZero() {
		fmt.Printf("✅ Successfully applied migrations: %s\n", group)

		// Record in database sequence
		fmt.Println(" Migration sequence updated in database")
	}
}

// revertLastMigration rolls back the last migration group
func revertLastMigration() {
	fmt.Println(" Reverting last migration (running down file)...")

	db, err := getDBConnection()
	if err != nil {
		log.Fatal("❌ ", err)
	}
	defer db.Close()

	migrator := migrate.NewMigrator(db, migrations.Migrations)

	// Lock migrations to prevent concurrent runs
	if err := migrator.Lock(context.Background()); err != nil {
		log.Fatalf("❌ Failed to lock migrations: %v", err)
	}
	defer migrator.Unlock(context.Background())

	// Get current status
	ms, err := migrator.MigrationsWithStatus(context.Background())
	if err != nil {
		log.Fatalf("❌ Failed to get migration status: %v", err)
	}

	lastGroup := ms.LastGroup()
	if lastGroup.IsZero() {
		fmt.Println("⚠️  No migrations to rollback")
		return
	}

	fmt.Printf(" Rolling back last group: %s\n", lastGroup)

	// Run rollback
	group, err := migrator.Rollback(context.Background())
	if err != nil {
		log.Fatalf("❌ Failed to rollback migration: %v", err)
	}

	if !group.IsZero() {
		fmt.Printf("✅ Successfully rolled back: %s\n", group)
		fmt.Println("📊 Migration sequence updated in database")
	}
}

// showMigrationStatus shows current migration status and sequence
func showMigrationStatus() {
	fmt.Println(" Checking migration status and sequence...")

	db, err := getDBConnection()
	if err != nil {
		log.Fatal("❌ ", err)
	}
	defer db.Close()

	migrator := migrate.NewMigrator(db, migrations.Migrations)

	// Check if migration table exists
	tableExists := true
	if err := migrator.Init(context.Background()); err != nil {
		tableExists = false
	}

	if !tableExists {
		fmt.Println("⚠️  Migration table not initialized. Run 'init' first.")
		return
	}

	// Get migration status
	ms, err := migrator.MigrationsWithStatus(context.Background())
	if err != nil {
		log.Fatalf("❌ Failed to get migration status: %v", err)
	}

	fmt.Printf(" Total migrations: %d\n", len(ms))
	fmt.Printf(" Applied migrations: %d\n", len(ms.Applied()))
	fmt.Printf(" Pending migrations: %d\n", len(ms.Unapplied()))

	if !ms.LastGroup().IsZero() {
		fmt.Printf(" Last migration group: %s\n", ms.LastGroup())
	}

	// Show migration sequence from database
	fmt.Println("\n Migration Sequence in Database:")

	var migrations []struct {
		ID         int64     `bun:"id"`
		Name       string    `bun:"name"`
		GroupID    int64     `bun:"group_id"`
		MigratedAt time.Time `bun:"migrated_at"`
	}

	err = db.NewSelect().
		Model(&migrations).
		Table("bun_migrations").
		Order("id ASC").
		Scan(context.Background())

	if err != nil {
		fmt.Printf("❌ Failed to read migration sequence: %v\n", err)
		return
	}

	if len(migrations) == 0 {
		fmt.Println("    No migrations applied yet")
	} else {
		for i, m := range migrations {
			fmt.Printf("   %d. %s (Group: %d, Applied: %s)\n",
				i+1, m.Name, m.GroupID, m.MigratedAt.Format("2006-01-02 15:04:05"))
		}
	}

	// Show pending migrations
	unapplied := ms.Unapplied()
	if len(unapplied) > 0 {
		fmt.Println("\n Pending Migrations:")
		for i, m := range unapplied {
			fmt.Printf("   %d. %s\n", i+1, m.Name)
		}
	}
}

// loadEnvFile loads environment variables from .env file if it exists
func loadEnvFile() {
	envFile := ".env"
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return // .env file doesn't exist, skip
	}

	file, err := os.Open(envFile)
	if err != nil {
		return // Can't open .env file, skip
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Invalid format
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Only set if environment variable is not already set
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

// runUnlock unlocks migrations when they get stuck
func runUnlock() {
	fmt.Println("🔓 Unlocking migrations...")

	db, err := getDBConnection()
	if err != nil {
		log.Fatal("❌ ", err)
	}
	defer db.Close()

	migrator := migrate.NewMigrator(db, migrations.Migrations)

	if err := migrator.Unlock(context.Background()); err != nil {
		log.Fatalf("❌ Failed to unlock migrations: %v", err)
	}

	fmt.Println("✅ Migrations unlocked successfully!")
}

// runVerify checks the database for any changes
func runVerify() {
	fmt.Println(" Verifying database changes...")

	db, err := getDBConnection()
	if err != nil {
		log.Fatal("❌ ", err)
	}
	defer db.Close()

	migrator := migrate.NewMigrator(db, migrations.Migrations)

	// Get current status
	ms, err := migrator.MigrationsWithStatus(context.Background())
	if err != nil {
		log.Fatalf("❌ Failed to get migration status: %v", err)
	}

	applied := ms.Applied()
	if len(applied) == 0 {
		fmt.Println("✅ No migrations have been applied yet")
		return
	}

	fmt.Printf(" Found %d applied migration(s)\n", len(applied))

	// Check each applied migration
	for _, m := range applied {
		fmt.Printf("   - %s\n", m.Name)
	}

	fmt.Println("✅ Verification completed")
}

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
