package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/vapusdata-ecosystem/vapusai/core/syntheticdata/generator"
)

// synthDataCmd represents the synth-data command
func NewSynthDataCmd() *cobra.Command {
	var (
		configFile     string
		dbSuffix       string
		rowCount       int
		parallelTables int
		dryRun         bool
		generateConfig string
		validateOnly   bool
	)

	synthDataCmd := &cobra.Command{
		Use:   "synth-data",
		Short: "Generate synthetic data for testing and development",
		Long: `Generate synthetic data by extracting schema from source PostgreSQL database,
creating realistic fake data, and loading it into a new test database.

This command performs a complete ETL (Extract, Transform, Load) process:
- Extract: Read table schemas from source database
- Transform: Generate synthetic data using go-faker with privacy rules
- Load: Create new database and populate with synthetic data

Examples:
  # Generate config template
  vapusctl synth-data --generate-config config.yaml
  
  # Generate synthetic data using config file
  vapusctl synth-data --config config.yaml
  
  # Generate with custom parameters
  vapusctl synth-data --config config.yaml --suffix "_test" --rows 5000
  
  # Dry run to preview schema only
  vapusctl synth-data --config config.yaml --dry-run
  
  # Validate connections only
  vapusctl synth-data --config config.yaml --validate-only`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Generate config template if requested
			if generateConfig != "" {
				return generateConfigTemplate(generateConfig)
			}

			// Validate that config file is provided
			if configFile == "" {
				return fmt.Errorf("config file is required. Use --config flag or --generate-config to create a template")
			}

			// Load configuration
			config, err := generator.LoadConfigFromFile(configFile)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Override config values from command line flags
			if dbSuffix != "" {
				config.DatabaseSuffix = dbSuffix
			}
			if rowCount > 0 {
				config.RowCount = rowCount
			}
			if parallelTables > 0 {
				config.ParallelTables = parallelTables
			}

			// Create generator instance
			sdg, err := generator.NewSyntheticDataGenerator(*config)
			if err != nil {
				return fmt.Errorf("failed to initialize synthetic data generator: %w", err)
			}
			defer sdg.Close()

			// Validate connections if requested
			if validateOnly {
				return validateConnections(sdg)
			}

			// Dry run to preview schema
			if dryRun {
				return previewSchema(sdg)
			}

			// Generate synthetic data
			return generateSyntheticData(sdg, config)
		},
	}

	// Add flags
	synthDataCmd.Flags().StringVarP(&configFile, "config", "c", "", "Configuration file path (required)")
	synthDataCmd.Flags().StringVarP(&dbSuffix, "suffix", "s", "", "Database suffix (overrides config)")
	synthDataCmd.Flags().IntVarP(&rowCount, "rows", "r", 0, "Number of rows to generate per table (overrides config)")
	synthDataCmd.Flags().IntVarP(&parallelTables, "parallel", "p", 0, "Number of tables to process in parallel (overrides config)")
	synthDataCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview schema without generating data")
	synthDataCmd.Flags().StringVar(&generateConfig, "generate-config", "", "Generate configuration template file")
	synthDataCmd.Flags().BoolVar(&validateOnly, "validate-only", false, "Only validate database connections")

	return synthDataCmd
}

// generateConfigTemplate creates a configuration template file
func generateConfigTemplate(filePath string) error {
	log.Printf("Generating configuration template: %s", filePath)

	if err := generator.GenerateConfigTemplate(filePath); err != nil {
		return fmt.Errorf("failed to generate config template: %w", err)
	}

	fmt.Printf("✓ Configuration template generated: %s\n", filePath)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Edit the configuration file with your database details")
	fmt.Println("2. Configure privacy rules for sensitive columns")
	fmt.Printf("3. Run: vapusctl synth-data --config %s\n", filePath)

	return nil
}

// validateConnections tests database connections
func validateConnections(sdg *generator.SyntheticDataGenerator) error {
	log.Printf("Validating database connections...")

	if err := sdg.ValidateConnection(); err != nil {
		return fmt.Errorf("connection validation failed: %w", err)
	}

	fmt.Println("✓ Source database connection: OK")
	fmt.Println("✓ Target database server connection: OK")
	fmt.Println("✓ All connections validated successfully")

	return nil
}

// previewSchema shows the extracted schema without generating data
func previewSchema(sdg *generator.SyntheticDataGenerator) error {
	log.Printf("Previewing database schema...")

	schema, err := sdg.PreviewSchema()
	if err != nil {
		return fmt.Errorf("failed to preview schema: %w", err)
	}

	fmt.Printf("\n Database Schema Preview: %s\n", schema.DatabaseName)
	fmt.Printf("─────────────────────────────────────────\n")
	fmt.Printf("Total tables: %d\n\n", len(schema.Tables))

	for i, table := range schema.Tables {
		fmt.Printf("%d. Table: %s (%d columns)\n", i+1, table.TableName, len(table.Columns))
		for j, column := range table.Columns {
			nullable := "NOT NULL"
			if column.IsNullable == "YES" {
				nullable = "NULL"
			}
			fmt.Printf("   %d. %s - %s (%s)\n", j+1, column.ColumnName, column.DataType, nullable)
		}
		fmt.Println()
	}

	fmt.Println(" Tip: Run without --dry-run to generate synthetic data")

	return nil
}

// generateSyntheticData performs the complete ETL process
func generateSyntheticData(sdg *generator.SyntheticDataGenerator, config *generator.GeneratorConfig) error {
	fmt.Printf(" Starting synthetic data generation\n")
	fmt.Printf("   Source DB: %s@%s:%d/%s\n", config.SourceDB.Username, config.SourceDB.Host, config.SourceDB.Port, config.SourceDB.Database)
	fmt.Printf("   Target DB: %s%s\n", config.SourceDB.Database, config.DatabaseSuffix)
	fmt.Printf("   Rows per table: %d\n", config.RowCount)
	fmt.Printf("   Parallel tables: %d\n", config.ParallelTables)
	fmt.Printf("   Privacy rules: %d\n\n", len(config.PrivacyRules))

	// Validate connections first
	if err := sdg.ValidateConnection(); err != nil {
		return fmt.Errorf("connection validation failed: %w", err)
	}
	fmt.Println("✓ Connections validated")

	// Generate synthetic data
	result, err := sdg.GenerateSyntheticData()
	if err != nil {
		return fmt.Errorf("synthetic data generation failed: %w", err)
	}

	// Display results
	fmt.Printf("\n Synthetic data generation completed successfully!\n")
	fmt.Printf("─────────────────────────────────────────────────────\n")
	fmt.Printf("Database: %s\n", result.DatabaseName)
	fmt.Printf("Generated at: %s\n", result.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Total tables processed: %d\n", len(result.Tables))

	totalRows := 0
	fmt.Printf("\nTable Summary:\n")
	for _, table := range result.Tables {
		rowCount := len(table.Rows)
		totalRows += rowCount
		fmt.Printf("  • %s: %d rows\n", table.TableName, rowCount)
	}
	fmt.Printf("\nTotal rows generated: %d\n", totalRows)

	// Save result to file
	resultFile := fmt.Sprintf("synthetic_data_result_%s.json", result.GeneratedAt.Format("20060102_150405"))
	if err := generator.SaveResultToFile(result, resultFile); err != nil {
		log.Printf("Warning: failed to save result file: %v", err)
	} else {
		fmt.Printf("Result saved to: %s\n", resultFile)
	}

	fmt.Printf("\n Tips:\n")
	fmt.Printf("  • Use the new database for testing: %s\n", result.DatabaseName)
	fmt.Printf("  • Check the result file for detailed generation info\n")
	fmt.Printf("  • Configure privacy rules in your config for sensitive data\n")

	return nil
}
