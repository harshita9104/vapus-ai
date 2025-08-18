package generator

import (
	"time"
)

// DatabaseConfig represents the configuration for database connections
type DatabaseConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Database string `json:"database" yaml:"database"`
	SSLMode  string `json:"sslmode" yaml:"sslmode"`
}

// ColumnInfo represents the metadata of a database column
type ColumnInfo struct {
	ColumnName   string  `json:"column_name"`
	DataType     string  `json:"data_type"`
	IsNullable   string  `json:"is_nullable"`
	DefaultValue *string `json:"column_default,omitempty"`
	MaxLength    int     `json:"character_maximum_length"`
}

// TableSchema represents the schema of a database table
type TableSchema struct {
	TableName string       `json:"table_name"`
	Columns   []ColumnInfo `json:"columns"`
}

// DatabaseSchema represents the complete schema of a database
type DatabaseSchema struct {
	DatabaseName string        `json:"database_name"`
	Tables       []TableSchema `json:"tables"`
}

// PrivacyRule represents rules for handling sensitive data
type PrivacyRule struct {
	TableName   string `json:"table_name"`
	ColumnName  string `json:"column_name"`
	FakerTag    string `json:"faker_tag"`
	Action      string `json:"action"` // "fake", "mask", "redact"
	CustomValue string `json:"custom_value,omitempty"`
}

// GeneratorConfig represents the configuration for the synthetic data generator
type GeneratorConfig struct {
	SourceDB       DatabaseConfig `json:"source_db" yaml:"source_db"`
	TargetDB       DatabaseConfig `json:"target_db" yaml:"target_db"`
	DatabaseSuffix string         `json:"database_suffix" yaml:"database_suffix"`
	RowCount       int            `json:"row_count" yaml:"row_count"`
	PrivacyRules   []PrivacyRule  `json:"privacy_rules" yaml:"privacy_rules"`
	ParallelTables int            `json:"parallel_tables" yaml:"parallel_tables"`
}

// GeneratedRow represents a row of generated synthetic data
type GeneratedRow map[string]interface{}

// TableData represents generated data for a table
type TableData struct {
	TableName string         `json:"table_name"`
	Rows      []GeneratedRow `json:"rows"`
}

// SyntheticDataResult represents the result of data generation
type SyntheticDataResult struct {
	DatabaseName string         `json:"database_name"`
	Tables       []TableData    `json:"tables"`
	GeneratedAt  time.Time      `json:"generated_at"`
	RowCounts    map[string]int `json:"row_counts"`
}

// DataGenerationStats represents statistics about the data generation process
type DataGenerationStats struct {
	TotalTables       int           `json:"total_tables"`
	TotalRows         int           `json:"total_rows"`
	ProcessingTime    time.Duration `json:"processing_time"`
	TablesProcessed   []string      `json:"tables_processed"`
	ErrorsEncountered []string      `json:"errors_encountered"`
}
