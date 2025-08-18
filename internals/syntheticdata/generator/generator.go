package generator

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// SyntheticDataGenerator is the main orchestrator for the ETL process
type SyntheticDataGenerator struct {
	config      GeneratorConfig
	extractor   *DatabaseExtractor
	transformer *DataTransformer
	loader      *DatabaseLoader
}

// NewSyntheticDataGenerator creates a new SyntheticDataGenerator instance
func NewSyntheticDataGenerator(config GeneratorConfig) (*SyntheticDataGenerator, error) {
	generator := &SyntheticDataGenerator{
		config: config,
	}

	// Initialize extractor
	extractor, err := NewDatabaseExtractor(config.SourceDB)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database extractor: %w", err)
	}
	generator.extractor = extractor

	// Initialize transformer
	generator.transformer = NewDataTransformer(config)

	// Initialize loader
	loader, err := NewDatabaseLoader(config.SourceDB, config.TargetDB)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database loader: %w", err)
	}
	generator.loader = loader

	return generator, nil
}

// Close cleans up all resources
func (sdg *SyntheticDataGenerator) Close() error {
	var errors []string

	if sdg.extractor != nil {
		if err := sdg.extractor.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("extractor: %v", err))
		}
	}

	if sdg.loader != nil {
		if err := sdg.loader.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("loader: %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during cleanup: %v", errors)
	}

	return nil
}

// GenerateSyntheticData performs the complete ETL process
func (sdg *SyntheticDataGenerator) GenerateSyntheticData() (*SyntheticDataResult, error) {
	log.Printf("Starting synthetic data generation process")
	startTime := time.Now()

	// Step 1: Extract schema from source database
	log.Printf("Step 1: Extracting schema from source database: %s", sdg.config.SourceDB.Database)
	schema, err := sdg.extractor.ExtractFullSchema()
	if err != nil {
		return nil, fmt.Errorf("failed to extract schema: %w", err)
	}
	log.Printf("Extracted schema for %d tables", len(schema.Tables))

	// Step 2: Create target database
	targetDBName := sdg.config.SourceDB.Database + sdg.config.DatabaseSuffix
	log.Printf("Step 2: Creating target database: %s", targetDBName)
	if err := sdg.loader.CreateDatabase(targetDBName); err != nil {
		return nil, fmt.Errorf("failed to create target database: %w", err)
	}

	// Step 3: Create table structures in target database
	log.Printf("Step 3: Creating table structures in target database")
	if err := sdg.loader.CreateTableStructure(schema); err != nil {
		return nil, fmt.Errorf("failed to create table structures: %w", err)
	}

	// Step 4: Generate synthetic data
	log.Printf("Step 4: Generating synthetic data")
	result, err := sdg.transformer.GenerateSyntheticData(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to generate synthetic data: %w", err)
	}

	// Step 5: Load synthetic data into target database
	log.Printf("Step 5: Loading synthetic data into target database")
	if err := sdg.loader.LoadSyntheticData(result); err != nil {
		return nil, fmt.Errorf("failed to load synthetic data: %w", err)
	}

	// Calculate processing time
	processingTime := time.Since(startTime)
	log.Printf("Synthetic data generation completed successfully in %v", processingTime)

	// Get final statistics
	stats, err := sdg.loader.GetTargetDatabaseStats()
	if err != nil {
		log.Printf("Warning: failed to get target database stats: %v", err)
	} else {
		log.Printf("Final statistics: %d tables, %d total rows", stats.TotalTables, stats.TotalRows)
	}

	return &result, nil
}

// ValidateConnection tests the database connections
func (sdg *SyntheticDataGenerator) ValidateConnection() error {
	log.Printf("Validating database connections")

	// Test source database connection
	if err := sdg.extractor.TestConnection(); err != nil {
		return fmt.Errorf("failed to connect to source database: %w", err)
	}
	log.Printf("Source database connection: OK")

	// Test target database connection (admin connection)
	adminConfig := sdg.config.TargetDB
	adminConfig.Database = "postgres"
	adminExtractor, err := NewDatabaseExtractor(adminConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to target database server: %w", err)
	}
	defer adminExtractor.Close()

	if err := adminExtractor.TestConnection(); err != nil {
		return fmt.Errorf("failed to connect to target database server: %w", err)
	}
	log.Printf("Target database server connection: OK")

	return nil
}

// PreviewSchema returns the extracted schema without generating data
func (sdg *SyntheticDataGenerator) PreviewSchema() (*DatabaseSchema, error) {
	log.Printf("Previewing database schema")
	schema, err := sdg.extractor.ExtractFullSchema()
	if err != nil {
		return nil, err
	}
	return &schema, nil
}

// GenerateConfigTemplate generates a configuration template file
func GenerateConfigTemplate(filePath string) error {
	template := GeneratorConfig{
		SourceDB: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Username: "postgres",
			Password: "password",
			Database: "source_database",
			SSLMode:  "disable",
		},
		TargetDB: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Username: "postgres",
			Password: "password",
			Database: "target_database",
			SSLMode:  "disable",
		},
		DatabaseSuffix: "_synthetic_test",
		RowCount:       1000,
		ParallelTables: 4,
		PrivacyRules: []PrivacyRule{
			{
				TableName:  "users",
				ColumnName: "email",
				FakerTag:   "email",
				Action:     "fake",
			},
			{
				TableName:  "users",
				ColumnName: "password",
				FakerTag:   "password",
				Action:     "fake",
			},
			{
				TableName:  "customers",
				ColumnName: "credit_card",
				FakerTag:   "credit_card",
				Action:     "fake",
			},
		},
	}

	// Determine file format based on extension
	ext := filepath.Ext(filePath)
	var data []byte
	var err error

	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(template)
	case ".json":
		data, err = json.MarshalIndent(template, "", "  ")
	default:
		return fmt.Errorf("unsupported file format: %s (use .yaml, .yml, or .json)", ext)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	log.Printf("Configuration template generated: %s", filePath)
	return nil
}

// LoadConfigFromFile loads configuration from a file
func LoadConfigFromFile(filePath string) (*GeneratorConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	var config GeneratorConfig
	ext := filepath.Ext(filePath)

	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &config)
	case ".json":
		err = json.Unmarshal(data, &config)
	default:
		return nil, fmt.Errorf("unsupported file format: %s (use .yaml, .yml, or .json)", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration file: %w", err)
	}

	// Set defaults
	if config.RowCount <= 0 {
		config.RowCount = 1000
	}
	if config.ParallelTables <= 0 {
		config.ParallelTables = 4
	}
	if config.DatabaseSuffix == "" {
		config.DatabaseSuffix = "_synthetic"
	}
	if config.SourceDB.SSLMode == "" {
		config.SourceDB.SSLMode = "disable"
	}
	if config.TargetDB.SSLMode == "" {
		config.TargetDB.SSLMode = "disable"
	}

	return &config, nil
}

// SaveResultToFile saves the generation result to a file
func SaveResultToFile(result *SyntheticDataResult, filePath string) error {
	ext := filepath.Ext(filePath)
	var data []byte
	var err error

	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(result)
	case ".json":
		data, err = json.MarshalIndent(result, "", "  ")
	default:
		return fmt.Errorf("unsupported file format: %s (use .yaml, .yml, or .json)", ext)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write result file: %w", err)
	}

	log.Printf("Result saved to: %s", filePath)
	return nil
}
