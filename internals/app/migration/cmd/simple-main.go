package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"

	// Import the existing migrations
	"github.com/vapusdata-ecosystem/vapusai/core/app/migration/migrations"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := os.Args[1]

	switch command {
	case "init":
		runInit()
	case "migrate":
		runMigrate()
	case "rollback":
		runRollback()
	case "status":
		runStatus()
	case "create-sql":
		if len(os.Args) < 3 {
			fmt.Println("Usage: vapus-migrate create-sql <migration_name>")
			return
		}
		createSQLMigration(strings.Join(os.Args[2:], "_"))
	case "create-go":
		if len(os.Args) < 3 {
			fmt.Println("Usage: vapus-migrate create-go <migration_name>")
			return
		}
		createGoMigration(strings.Join(os.Args[2:], "_"))
	case "mark-applied":
		runMarkApplied()
	case "unlock":
		runUnlock()
	case "verify":
		runVerify()
	case "delete-user":
		if len(os.Args) < 3 {
			fmt.Println("Usage: vapus-migrate delete-user <email>")
			return
		}
		deleteUser(os.Args[2])
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
	}
}

func printHelp() {
	fmt.Println(`VapusAI Database Migration Tool

Usage:
  vapus-migrate <command>

Commands:
  init        Initialize migration table
  migrate     Run pending migrations
  rollback    Rollback last migration group
  status      Show migration status
  mark-applied Mark migrations as applied without running them
  unlock      Unlock migrations (use if locked)
  create-sql  Create new SQL migration file
  create-go   Create new Go migration file
  verify      Check database changes
  delete-user Delete a user by email
  help        Show this help

Environment Variables:
  DB_HOST     Database host (default: localhost)
  DB_PORT     Database port (default: 5432)
  DB_USER     Database username (default: postgres)
  DB_PASSWORD Database password (required)
  DB_NAME     Database name (default: vapusai)
  DB_SSLMODE  SSL mode (default: disable)

Examples:
  vapus-migrate init
  vapus-migrate migrate
  vapus-migrate mark-applied
  vapus-migrate create-sql add_user_column
  vapus-migrate create-go update_vendor_tables`)
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

func runMigrate() {
	fmt.Println(" Running pending migrations...")

	db, err := getDBConnection()
	if err != nil {
		log.Fatal("❌ ", err)
	}
	defer db.Close()

	migrator := migrate.NewMigrator(db, migrations.Migrations)

	if err := migrator.Lock(context.Background()); err != nil {
		log.Fatal("❌ Failed to lock migrations: ", err)
	}
	defer migrator.Unlock(context.Background())

	group, err := migrator.Migrate(context.Background())
	if err != nil {
		log.Fatal("❌ Failed to run migrations: ", err)
	}

	if group.IsZero() {
		fmt.Println("✅ No new migrations to run (database is up to date)")
	} else {
		fmt.Printf("✅ Successfully migrated to: %s\n", group)
	}
}

func runRollback() {
	fmt.Println(" Rolling back last migration group...")

	db, err := getDBConnection()
	if err != nil {
		log.Fatal("❌ ", err)
	}
	defer db.Close()

	migrator := migrate.NewMigrator(db, migrations.Migrations)

	if err := migrator.Lock(context.Background()); err != nil {
		log.Fatal("❌ Failed to lock migrations: ", err)
	}
	defer migrator.Unlock(context.Background())

	group, err := migrator.Rollback(context.Background())
	if err != nil {
		log.Fatal("❌ Failed to rollback migrations: ", err)
	}

	if group.IsZero() {
		fmt.Println("✅ No migration groups to roll back")
	} else {
		fmt.Printf("✅ Successfully rolled back: %s\n", group)
	}
}

func runStatus() {
	fmt.Println("📊 Checking migration status...")

	db, err := getDBConnection()
	if err != nil {
		log.Fatal("❌ ", err)
	}
	defer db.Close()

	migrator := migrate.NewMigrator(db, migrations.Migrations)

	ms, err := migrator.MigrationsWithStatus(context.Background())
	if err != nil {
		log.Fatal("❌ Failed to get migration status: ", err)
	}

	fmt.Printf("📋 All migrations: %s\n", ms)
	fmt.Printf("⏳ Unapplied migrations: %s\n", ms.Unapplied())
	fmt.Printf("✅ Last migration group: %s\n", ms.LastGroup())
}

func runMarkApplied() {
	fmt.Println("📝 Marking migrations as applied without running them...")

	db, err := getDBConnection()
	if err != nil {
		log.Fatal("❌ ", err)
	}
	defer db.Close()

	migrator := migrate.NewMigrator(db, migrations.Migrations)

	if err := migrator.Lock(context.Background()); err != nil {
		log.Fatal("❌ Failed to lock migrations: ", err)
	}
	defer migrator.Unlock(context.Background())

	group, err := migrator.Migrate(context.Background(), migrate.WithNopMigration())
	if err != nil {
		log.Fatal("❌ Failed to mark migrations as applied: ", err)
	}

	if group.IsZero() {
		fmt.Println("✅ No new migrations to mark as applied")
	} else {
		fmt.Printf("✅ Successfully marked as applied: %s\n", group)
	}
}

func runUnlock() {
	fmt.Println("🔓 Unlocking migrations...")

	db, err := getDBConnection()
	if err != nil {
		log.Fatal("❌ ", err)
	}
	defer db.Close()

	migrator := migrate.NewMigrator(db, migrations.Migrations)

	if err := migrator.Unlock(context.Background()); err != nil {
		log.Fatal("❌ Failed to unlock migrations: ", err)
	}

	fmt.Println("✅ Migrations unlocked successfully!")
}

func createSQLMigration(name string) {
	fmt.Printf(" Creating SQL migration: %s\n", name)

	db, err := getDBConnection()
	if err != nil {
		log.Fatal("❌ ", err)
	}
	defer db.Close()

	migrator := migrate.NewMigrator(db, migrations.Migrations)

	files, err := migrator.CreateSQLMigrations(context.Background(), name)
	if err != nil {
		log.Fatal("❌ Failed to create SQL migration: ", err)
	}

	for _, mf := range files {
		fmt.Printf("✅ Created migration: %s (%s)\n", mf.Name, mf.Path)
	}
}

func createGoMigration(name string) {
	fmt.Printf(" Creating Go migration: %s\n", name)

	db, err := getDBConnection()
	if err != nil {
		log.Fatal("❌ ", err)
	}
	defer db.Close()

	migrator := migrate.NewMigrator(db, migrations.Migrations)

	mf, err := migrator.CreateGoMigration(context.Background(), name)
	if err != nil {
		log.Fatal("❌ Failed to create Go migration: ", err)
	}

	fmt.Printf("✅ Created migration: %s (%s)\n", mf.Name, mf.Path)
}

func runVerify() {
	fmt.Println("🔍 VERIFYING DATABASE CHANGES")
	fmt.Println("==================================================")

	db, err := getDBConnection()
	if err != nil {
		log.Fatal("❌ ", err)
	}
	defer db.Close()

	// Check if migration_test_table exists
	var tableExists bool
	err = db.NewSelect().
		ColumnExpr("true").
		TableExpr("information_schema.tables").
		Where("table_name = ? AND table_schema = 'public'", "migration_test_table").
		Scan(context.Background(), &tableExists)

	if err != nil && err.Error() != "sql: no rows in result set" {
		log.Fatalf("Error checking table existence: %v", err)
	}

	if tableExists {
		fmt.Println("✅ migration_test_table exists")

		// Count records in the test table
		var count int
		err = db.NewSelect().
			ColumnExpr("count(*)").
			TableExpr("migration_test_table").
			Scan(context.Background(), &count)

		if err != nil {
			log.Fatalf("Error counting records: %v", err)
		}

		fmt.Printf("✅ migration_test_table has %d record(s)\n", count)

		// Show the records
		var records []struct {
			ID   int    `bun:"id"`
			Name string `bun:"name"`
		}

		err = db.NewSelect().
			Model(&records).
			Table("migration_test_table").
			Scan(context.Background())

		if err != nil {
			log.Fatalf("Error fetching records: %v", err)
		}

		fmt.Printf("📋 Records in migration_test_table:\n")
		for _, record := range records {
			fmt.Printf("   ID: %d, Name: %s\n", record.ID, record.Name)
		}
	} else {
		fmt.Println("❌ migration_test_table does not exist")
	}

	// Check bun_migrations table
	var migrationCount int
	err = db.NewSelect().
		ColumnExpr("count(*)").
		TableExpr("bun_migrations").
		Scan(context.Background(), &migrationCount)

	if err != nil {
		log.Fatalf("Error counting migrations: %v", err)
	}

	fmt.Printf("✅ bun_migrations table has %d migration(s) recorded\n", migrationCount)

	// Check users table
	var userCount int
	err = db.NewSelect().
		ColumnExpr("count(*)").
		TableExpr("users").
		Scan(context.Background(), &userCount)

	if err != nil {
		fmt.Printf("❌ Could not access users table: %v\n", err)
	} else {
		fmt.Printf("✅ users table has %d user(s)\n", userCount)
	}

	fmt.Println("==================================================")
	fmt.Println("✅ Database verification complete!")
}

func deleteUser(email string) {
	fmt.Printf("🗑️  DELETING USER WITH EMAIL: %s\n", email)
	fmt.Println("==================================================")

	db, err := getDBConnection()
	if err != nil {
		log.Fatal("❌ ", err)
	}
	defer db.Close()

	// First, check if user exists
	var userExists bool
	err = db.NewSelect().
		ColumnExpr("true").
		TableExpr("users").
		Where("user_id = ?", email).
		Scan(context.Background(), &userExists)

	if err != nil && err.Error() != "sql: no rows in result set" {
		log.Fatalf("Error checking user existence: %v", err)
	}

	if !userExists {
		fmt.Printf("❌ User with email '%s' not found\n", email)
		return
	}

	// Get user details before deletion
	var user struct {
		DisplayName string `bun:"display_name"`
		UserID      string `bun:"user_id"`
	}

	err = db.NewSelect().
		Model(&user).
		Table("users").
		Where("user_id = ?", email).
		Scan(context.Background())

	if err != nil {
		log.Fatalf("Error getting user details: %v", err)
	}

	fmt.Printf("👤 Found user: %s (%s)\n", user.DisplayName, user.UserID)

	// Delete the user
	result, err := db.NewDelete().
		Table("users").
		Where("user_id = ?", email).
		Exec(context.Background())

	if err != nil {
		log.Fatalf("Error deleting user: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatalf("Error getting affected rows: %v", err)
	}

	if rowsAffected > 0 {
		fmt.Printf("✅ Successfully deleted %d user(s) with email '%s'\n", rowsAffected, email)
	} else {
		fmt.Printf("❌ No users were deleted with email '%s'\n", email)
	}

	fmt.Println("==================================================")
}
