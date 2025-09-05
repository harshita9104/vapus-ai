package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration: dynamic tables - add status column] ")
		
		// This migration adds a 'status' column to all vendor tables
		// Get all vendor tables (tables that start with 'vendor_')
		rows, err := db.QueryContext(ctx, `
			SELECT tablename FROM pg_tables 
			WHERE tablename LIKE 'vendor_%' 
			AND schemaname = 'public'
		`)
		if err != nil {
			return fmt.Errorf("failed to query vendor tables: %w", err)
		}
		defer rows.Close()

		var tableNames []string
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				return fmt.Errorf("failed to scan table name: %w", err)
			}
			tableNames = append(tableNames, tableName)
		}

		// Add status column to each vendor table
		for _, tableName := range tableNames {
			query := fmt.Sprintf(`
				ALTER TABLE %s 
				ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active'
			`, tableName)
			
			_, err := db.ExecContext(ctx, query)
			if err != nil {
				return fmt.Errorf("failed to add status column to table %s: %w", tableName, err)
			}
			fmt.Printf(" [added status column to %s] ", tableName)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration: dynamic tables - remove status column] ")
		
		// Rollback: Remove status column from all vendor tables
		rows, err := db.QueryContext(ctx, `
			SELECT tablename FROM pg_tables 
			WHERE tablename LIKE 'vendor_%' 
			AND schemaname = 'public'
		`)
		if err != nil {
			return fmt.Errorf("failed to query vendor tables: %w", err)
		}
		defer rows.Close()

		var tableNames []string
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				return fmt.Errorf("failed to scan table name: %w", err)
			}
			tableNames = append(tableNames, tableName)
		}

		// Remove status column from each vendor table
		for _, tableName := range tableNames {
			query := fmt.Sprintf(`
				ALTER TABLE %s 
				DROP COLUMN IF EXISTS status
			`, tableName)
			
			_, err := db.ExecContext(ctx, query)
			if err != nil {
				return fmt.Errorf("failed to remove status column from table %s: %w", tableName, err)
			}
			fmt.Printf(" [removed status column from %s] ", tableName)
		}

		return nil
	})
}
