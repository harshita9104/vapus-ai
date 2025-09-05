package migrations

import (
	"context"
	"fmt"
	"strings"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration: adding column to dynamic vendor tables] ")
		
		// Get all vendor tables (tables that start with 'vendor_')
		var tables []string
		err := db.NewRaw(`
			SELECT table_name 
			FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name LIKE 'vendor_%'
		`).Scan(ctx, &tables)
		
		if err != nil {
			return fmt.Errorf("failed to get vendor tables: %w", err)
		}

		// Add column to each vendor table
		for _, table := range tables {
			// Sanitize table name to prevent SQL injection
			if !isValidTableName(table) {
				continue
			}
			
			query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP", table)
			_, err := db.ExecContext(ctx, query)
			if err != nil {
				return fmt.Errorf("failed to add column to table %s: %w", table, err)
			}
			fmt.Printf("Added last_updated column to %s ", table)
		}
		
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration: removing column from dynamic vendor tables] ")
		
		// Get all vendor tables
		var tables []string
		err := db.NewRaw(`
			SELECT table_name 
			FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name LIKE 'vendor_%'
		`).Scan(ctx, &tables)
		
		if err != nil {
			return fmt.Errorf("failed to get vendor tables: %w", err)
		}

		// Remove column from each vendor table
		for _, table := range tables {
			if !isValidTableName(table) {
				continue
			}
			
			query := fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS last_updated", table)
			_, err := db.ExecContext(ctx, query)
			if err != nil {
				return fmt.Errorf("failed to remove column from table %s: %w", table, err)
			}
			fmt.Printf("Removed last_updated column from %s ", table)
		}
		
		return nil
	})
}

// isValidTableName validates table name to prevent SQL injection
func isValidTableName(name string) bool {
	// Allow only alphanumeric characters and underscores
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '_') {
			return false
		}
	}
	return len(name) > 0 && len(name) <= 64 && strings.HasPrefix(name, "vendor_")
}
