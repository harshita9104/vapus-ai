package generator

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// DatabaseExtractor handles extracting schema information from databases
type DatabaseExtractor struct {
	config DatabaseConfig
	db     *sql.DB
}

// NewDatabaseExtractor creates a new DatabaseExtractor instance
func NewDatabaseExtractor(config DatabaseConfig) (*DatabaseExtractor, error) {
	extractor := &DatabaseExtractor{
		config: config,
	}

	if err := extractor.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return extractor, nil
}

// connect establishes a connection to the database
func (de *DatabaseExtractor) connect() error {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		de.config.Host, de.config.Port, de.config.Username, de.config.Password, de.config.Database, de.config.SSLMode)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	de.db = db
	log.Printf("Successfully connected to database: %s", de.config.Database)
	return nil
}

// Close closes the database connection
func (de *DatabaseExtractor) Close() error {
	if de.db != nil {
		return de.db.Close()
	}
	return nil
}

// GetTableNames retrieves all table names from the public schema
func (de *DatabaseExtractor) GetTableNames() ([]string, error) {
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := de.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table names: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over table names: %w", err)
	}

	log.Printf("Found %d tables in database", len(tables))
	return tables, nil
}

// GetTableSchema retrieves the schema information for a specific table
func (de *DatabaseExtractor) GetTableSchema(tableName string) (TableSchema, error) {
	query := `
		SELECT
			a.attname AS column_name,
			pg_catalog.format_type(a.atttypid, a.atttypmod) AS data_type,
			CASE WHEN a.attnotnull THEN 'NO' ELSE 'YES' END AS is_nullable,
			pg_catalog.pg_get_expr(d.adbin, d.adrelid) AS column_default,
			CASE 
				WHEN a.atttypmod > 0 AND pg_catalog.format_type(a.atttypid, a.atttypmod) LIKE '%varying%' 
				THEN a.atttypmod - 4 
				ELSE NULL 
			END AS character_maximum_length
		FROM
			pg_catalog.pg_attribute a
		JOIN
			pg_catalog.pg_class c ON a.attrelid = c.oid
		JOIN
			pg_catalog.pg_namespace n ON c.relnamespace = n.oid
		LEFT JOIN
			pg_catalog.pg_attrdef d ON a.attrelid = d.adrelid AND a.attnum = d.adnum
		WHERE
			c.relname = $1
			AND n.nspname = 'public'
			AND a.attnum > 0
			AND NOT a.attisdropped
		ORDER BY
			a.attnum
	`

	rows, err := de.db.Query(query, tableName)
	if err != nil {
		return TableSchema{}, fmt.Errorf("failed to query table schema for %s: %w", tableName, err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var maxLength sql.NullInt64
		var defaultValue sql.NullString

		if err := rows.Scan(&col.ColumnName, &col.DataType, &col.IsNullable, &defaultValue, &maxLength); err != nil {
			return TableSchema{}, fmt.Errorf("failed to scan column info: %w", err)
		}

		if maxLength.Valid {
			col.MaxLength = int(maxLength.Int64)
		}

		if defaultValue.Valid {
			col.DefaultValue = &defaultValue.String
		}

		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return TableSchema{}, fmt.Errorf("error iterating over columns: %w", err)
	}

	return TableSchema{
		TableName: tableName,
		Columns:   columns,
	}, nil
}

// ExtractFullSchema extracts the complete database schema
func (de *DatabaseExtractor) ExtractFullSchema() (DatabaseSchema, error) {
	log.Printf("Starting schema extraction for database: %s", de.config.Database)

	tableNames, err := de.GetTableNames()
	if err != nil {
		return DatabaseSchema{}, fmt.Errorf("failed to get table names: %w", err)
	}

	var tables []TableSchema
	for _, tableName := range tableNames {
		tableSchema, err := de.GetTableSchema(tableName)
		if err != nil {
			log.Printf("Warning: failed to get schema for table %s: %v", tableName, err)
			continue
		}
		tables = append(tables, tableSchema)
		log.Printf("Extracted schema for table: %s (%d columns)", tableName, len(tableSchema.Columns))
	}

	schema := DatabaseSchema{
		DatabaseName: de.config.Database,
		Tables:       tables,
	}

	log.Printf("Schema extraction completed. Found %d tables with schemas", len(tables))
	return schema, nil
}

// GetTableRowCount returns the current row count for a table
func (de *DatabaseExtractor) GetTableRowCount(tableName string) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)

	var count int
	err := de.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get row count for table %s: %w", tableName, err)
	}

	return count, nil
}

// TestConnection tests if the database connection is working
func (de *DatabaseExtractor) TestConnection() error {
	return de.db.Ping()
}
