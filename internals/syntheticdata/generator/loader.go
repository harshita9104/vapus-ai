package generator

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	_ "github.com/lib/pq"
)

// DatabaseLoader handles creating databases and loading synthetic data
type DatabaseLoader struct {
	sourceConfig DatabaseConfig
	targetConfig DatabaseConfig
	adminDB      *sql.DB
	targetDB     *sql.DB
}

// NewDatabaseLoader creates a new DatabaseLoader instance
func NewDatabaseLoader(sourceConfig, targetConfig DatabaseConfig) (*DatabaseLoader, error) {
	loader := &DatabaseLoader{
		sourceConfig: sourceConfig,
		targetConfig: targetConfig,
	}

	// Connect to admin database (typically 'postgres' database) for creating new databases
	adminConfig := targetConfig
	adminConfig.Database = "postgres"

	if err := loader.connectAdmin(adminConfig); err != nil {
		return nil, fmt.Errorf("failed to connect to admin database: %w", err)
	}

	return loader, nil
}

// connectAdmin connects to the admin database for administrative operations
func (dl *DatabaseLoader) connectAdmin(config DatabaseConfig) error {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	dl.adminDB = db
	log.Printf("Connected to admin database for administrative operations")
	return nil
}

// connectTarget connects to the target database for data operations
func (dl *DatabaseLoader) connectTarget() error {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dl.targetConfig.Host, dl.targetConfig.Port, dl.targetConfig.Username, dl.targetConfig.Password, dl.targetConfig.Database, dl.targetConfig.SSLMode)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	dl.targetDB = db
	log.Printf("Connected to target database: %s", dl.targetConfig.Database)

	// Enable required PostgreSQL extensions
	if err := dl.enableExtensions(); err != nil {
		return fmt.Errorf("failed to enable database extensions: %w", err)
	}

	return nil
}

// Close closes all database connections
func (dl *DatabaseLoader) Close() error {
	var errors []string

	if dl.adminDB != nil {
		if err := dl.adminDB.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("admin DB: %v", err))
		}
	}

	if dl.targetDB != nil {
		if err := dl.targetDB.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("target DB: %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing connections: %s", strings.Join(errors, ", "))
	}

	return nil
}

// CreateDatabase creates a new database with the specified name
func (dl *DatabaseLoader) CreateDatabase(databaseName string) error {
	log.Printf("Creating database: %s", databaseName)

	// Check if database already exists
	checkQuery := "SELECT 1 FROM pg_database WHERE datname = $1"
	var exists int
	err := dl.adminDB.QueryRow(checkQuery, databaseName).Scan(&exists)
	if err == nil {
		log.Printf("Database %s already exists, dropping it first", databaseName)
		if err := dl.DropDatabase(databaseName); err != nil {
			return fmt.Errorf("failed to drop existing database: %w", err)
		}
	} else if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Create the new database
	createQuery := fmt.Sprintf("CREATE DATABASE %s", databaseName)
	_, err = dl.adminDB.Exec(createQuery)
	if err != nil {
		return fmt.Errorf("failed to create database %s: %w", databaseName, err)
	}

	log.Printf("Successfully created database: %s", databaseName)

	// Update target config to point to new database
	dl.targetConfig.Database = databaseName

	// Connect to the newly created database
	if err := dl.connectTarget(); err != nil {
		return fmt.Errorf("failed to connect to newly created database: %w", err)
	}

	return nil
}

// DropDatabase drops the specified database if it exists
func (dl *DatabaseLoader) DropDatabase(databaseName string) error {
	log.Printf("Dropping database: %s", databaseName)

	// Terminate all connections to the database
	terminateQuery := `
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = $1 AND pid <> pg_backend_pid()
	`
	_, err := dl.adminDB.Exec(terminateQuery, databaseName)
	if err != nil {
		log.Printf("Warning: failed to terminate connections to database %s: %v", databaseName, err)
	}

	// Drop the database
	dropQuery := fmt.Sprintf("DROP DATABASE IF EXISTS %s", databaseName)
	_, err = dl.adminDB.Exec(dropQuery)
	if err != nil {
		return fmt.Errorf("failed to drop database %s: %w", databaseName, err)
	}

	log.Printf("Successfully dropped database: %s", databaseName)
	return nil
}

// CreateTableStructure creates table structures in the target database
func (dl *DatabaseLoader) CreateTableStructure(schema DatabaseSchema) error {
	log.Printf("Creating table structures for %d tables", len(schema.Tables))

	// First pass: Create sequences for auto-incrementing columns
	err := dl.createSequences(schema.Tables)
	if err != nil {
		return fmt.Errorf("failed to create sequences: %w", err)
	}

	// Second pass: Create tables
	for _, table := range schema.Tables {
		if err := dl.createSingleTable(table); err != nil {
			return fmt.Errorf("failed to create table %s: %w", table.TableName, err)
		}
	}

	log.Printf("Successfully created all table structures")
	return nil
}

// createSingleTable creates a single table based on its schema
func (dl *DatabaseLoader) createSingleTable(tableSchema TableSchema) error {
	log.Printf("Creating table: %s", tableSchema.TableName)

	// Build CREATE TABLE statement
	var columns []string
	for _, column := range tableSchema.Columns {
		columnDef := fmt.Sprintf("%s %s", column.ColumnName, dl.mapDataType(column.DataType))

		if column.IsNullable == "NO" {
			columnDef += " NOT NULL"
		}

		if column.DefaultValue != nil && *column.DefaultValue != "" && *column.DefaultValue != "NULL" {
			columnDef += fmt.Sprintf(" DEFAULT %s", *column.DefaultValue)
		}

		columns = append(columns, columnDef)
	}

	createQuery := fmt.Sprintf("CREATE TABLE %s (\n    %s\n)",
		tableSchema.TableName,
		strings.Join(columns, ",\n    "))

	_, err := dl.targetDB.Exec(createQuery)
	if err != nil {
		return fmt.Errorf("failed to execute CREATE TABLE for %s: %w", tableSchema.TableName, err)
	}

	log.Printf("Successfully created table: %s", tableSchema.TableName)
	return nil
}

// mapDataType maps PostgreSQL data types for CREATE TABLE statements
func (dl *DatabaseLoader) mapDataType(dataType string) string {
	// Handle special PostgreSQL data types
	switch dataType {
	case "ARRAY":
		// Default ARRAY type to text[] for synthetic data
		return "text[]"
	case "USER-DEFINED":
		// Default user-defined types to text for synthetic data
		return "text"
	case "tsvector":
		// PostgreSQL full-text search vector type
		return "tsvector"
	case "character varying":
		return "varchar(255)"
	case "bigint":
		return "bigint"
	case "integer":
		return "integer"
	case "smallint":
		return "smallint"
	case "real":
		return "real"
	case "double precision":
		return "double precision"
	case "boolean":
		return "boolean"
	case "text":
		return "text"
	case "jsonb":
		return "jsonb"
	case "json":
		return "json"
	case "uuid":
		return "uuid"
	case "timestamp with time zone":
		return "timestamp with time zone"
	case "timestamp without time zone":
		return "timestamp without time zone"
	case "date":
		return "date"
	case "time":
		return "time"
	default:
		// For other types, use as-is but log for debugging
		log.Printf("Using data type as-is: %s", dataType)
		return dataType
	}
}

// LoadSyntheticData loads the generated synthetic data into the target database
func (dl *DatabaseLoader) LoadSyntheticData(result SyntheticDataResult) error {
	log.Printf("Loading synthetic data for %d tables", len(result.Tables))

	// Process tables concurrently
	semaphore := make(chan struct{}, 4) // Limit concurrent table operations
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []string

	for _, tableData := range result.Tables {
		wg.Add(1)
		go func(td TableData) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			if err := dl.loadTableData(td); err != nil {
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Table %s: %v", td.TableName, err))
				mu.Unlock()
				log.Printf("Error loading data for table %s: %v", td.TableName, err)
			} else {
				log.Printf("Successfully loaded %d rows into table %s", len(td.Rows), td.TableName)
			}
		}(tableData)
	}

	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("errors loading data: %s", strings.Join(errors, "; "))
	}

	log.Printf("Successfully loaded synthetic data for all tables")
	return nil
}

// loadTableData loads data for a single table using batch inserts
func (dl *DatabaseLoader) loadTableData(tableData TableData) error {
	if len(tableData.Rows) == 0 {
		log.Printf("No data to load for table %s", tableData.TableName)
		return nil
	}

	log.Printf("Loading %d rows into table: %s", len(tableData.Rows), tableData.TableName)

	// Get column names from the first row
	var columns []string
	for column := range tableData.Rows[0] {
		columns = append(columns, column)
	}

	// Build INSERT statement
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableData.TableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	// Prepare statement for better performance
	stmt, err := dl.targetDB.Prepare(insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	// Insert rows in batches
	batchSize := 1000
	for i := 0; i < len(tableData.Rows); i += batchSize {
		end := i + batchSize
		if end > len(tableData.Rows) {
			end = len(tableData.Rows)
		}

		if err := dl.insertBatch(stmt, tableData.Rows[i:end], columns); err != nil {
			return fmt.Errorf("failed to insert batch %d-%d: %w", i, end, err)
		}

		if i%5000 == 0 && i > 0 {
			log.Printf("Inserted %d/%d rows for table %s", i, len(tableData.Rows), tableData.TableName)
		}
	}

	return nil
}

// insertBatch inserts a batch of rows using the prepared statement
func (dl *DatabaseLoader) insertBatch(stmt *sql.Stmt, rows []GeneratedRow, columns []string) error {
	// Begin transaction for batch insert
	tx, err := dl.targetDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	txStmt := tx.Stmt(stmt)
	defer txStmt.Close()

	for _, row := range rows {
		values := make([]interface{}, len(columns))
		for i, column := range columns {
			values[i] = row[column]
		}

		_, err := txStmt.Exec(values...)
		if err != nil {
			return fmt.Errorf("failed to execute insert: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// createSequences creates PostgreSQL sequences for auto-incrementing columns
func (dl *DatabaseLoader) createSequences(tables []TableSchema) error {
	log.Printf("Creating sequences for auto-incrementing columns")

	for _, table := range tables {
		for _, column := range table.Columns {
			if column.DefaultValue != nil && *column.DefaultValue != "" {
				// Check if the default value references a sequence
				if strings.Contains(*column.DefaultValue, "nextval(") {
					sequenceName := dl.extractSequenceName(*column.DefaultValue)
					if sequenceName != "" {
						err := dl.createSequence(sequenceName, column.DataType)
						if err != nil {
							return fmt.Errorf("failed to create sequence %s: %w", sequenceName, err)
						}
					}
				}
			}
		}
	}

	return nil
}

// extractSequenceName extracts sequence name from a DEFAULT expression like "nextval('accounts_id_seq'::regclass)"
func (dl *DatabaseLoader) extractSequenceName(defaultValue string) string {
	// Pattern: nextval('sequence_name'::regclass)
	re := regexp.MustCompile(`nextval\('([^']+)'::regclass\)`)
	matches := re.FindStringSubmatch(defaultValue)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// createSequence creates a PostgreSQL sequence
func (dl *DatabaseLoader) createSequence(sequenceName, dataType string) error {
	log.Printf("Creating sequence: %s", sequenceName)

	// Determine start value based on data type
	startValue := 1
	if strings.Contains(dataType, "bigint") {
		startValue = 1
	}

	createSeqQuery := fmt.Sprintf(`
		CREATE SEQUENCE IF NOT EXISTS %s
		INCREMENT BY 1
		START WITH %d
		NO MINVALUE
		NO MAXVALUE
		CACHE 1;
	`, sequenceName, startValue)

	_, err := dl.targetDB.Exec(createSeqQuery)
	if err != nil {
		return fmt.Errorf("failed to execute CREATE SEQUENCE for %s: %w", sequenceName, err)
	}

	log.Printf("Successfully created sequence: %s", sequenceName)
	return nil
}

// enableExtensions enables required PostgreSQL extensions in the target database
func (dl *DatabaseLoader) enableExtensions() error {
	log.Printf("Enabling required PostgreSQL extensions")

	// Enable pgvector extension for vector data types
	_, err := dl.targetDB.Exec(`CREATE EXTENSION IF NOT EXISTS "vector"`)
	if err != nil {
		log.Printf("Warning: failed to enable vector extension (this is normal if pgvector is not installed): %v", err)
		// Don't return error here as vector extension might not be available on all systems
		// The error will be caught later when creating tables with vector columns
	} else {
		log.Printf("Successfully enabled vector extension")
	}

	return nil
}

// GetTargetDatabaseStats returns statistics about the target database
func (dl *DatabaseLoader) GetTargetDatabaseStats() (DataGenerationStats, error) {
	stats := DataGenerationStats{
		TablesProcessed: []string{},
	}

	// Get list of tables
	tableQuery := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
	`

	rows, err := dl.targetDB.Query(tableQuery)
	if err != nil {
		return stats, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	totalRows := 0
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}

		// Get row count for each table
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
		var count int
		if err := dl.targetDB.QueryRow(countQuery).Scan(&count); err == nil {
			totalRows += count
			stats.TablesProcessed = append(stats.TablesProcessed, tableName)
		}
	}

	stats.TotalTables = len(stats.TablesProcessed)
	stats.TotalRows = totalRows

	return stats, nil
}
