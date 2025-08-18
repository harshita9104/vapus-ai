package generator

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
)

// DataTransformer handles the transformation of schema into synthetic data
type DataTransformer struct {
	config       GeneratorConfig
	privacyRules map[string]map[string]PrivacyRule // table.column -> rule
}

// NewDataTransformer creates a new DataTransformer instance
func NewDataTransformer(config GeneratorConfig) *DataTransformer {
	dt := &DataTransformer{
		config:       config,
		privacyRules: make(map[string]map[string]PrivacyRule),
	}

	// Index privacy rules for quick lookup
	for _, rule := range config.PrivacyRules {
		if dt.privacyRules[rule.TableName] == nil {
			dt.privacyRules[rule.TableName] = make(map[string]PrivacyRule)
		}
		dt.privacyRules[rule.TableName][rule.ColumnName] = rule
	}

	return dt
}

// GenerateTableData generates synthetic data for a single table
func (dt *DataTransformer) GenerateTableData(tableSchema TableSchema, rowCount int) (TableData, error) {
	log.Printf("Generating %d rows for table: %s", rowCount, tableSchema.TableName)

	var rows []GeneratedRow
	for i := 0; i < rowCount; i++ {
		row, err := dt.generateRow(tableSchema)
		if err != nil {
			return TableData{}, fmt.Errorf("failed to generate row %d for table %s: %w", i, tableSchema.TableName, err)
		}
		rows = append(rows, row)
	}

	return TableData{
		TableName: tableSchema.TableName,
		Rows:      rows,
	}, nil
}

// generateRow generates a single row of synthetic data based on table schema
func (dt *DataTransformer) generateRow(tableSchema TableSchema) (GeneratedRow, error) {
	row := make(GeneratedRow)

	for _, column := range tableSchema.Columns {
		value, err := dt.generateColumnValue(tableSchema.TableName, column)
		if err != nil {
			return nil, fmt.Errorf("failed to generate value for column %s: %w", column.ColumnName, err)
		}
		row[column.ColumnName] = value
	}

	return row, nil
}

// generateColumnValue generates a value for a specific column based on its type and privacy rules
func (dt *DataTransformer) generateColumnValue(tableName string, column ColumnInfo) (interface{}, error) {
	// Check if there's a privacy rule for this column
	if tableRules, exists := dt.privacyRules[tableName]; exists {
		if rule, exists := tableRules[column.ColumnName]; exists {
			return dt.applyPrivacyRule(rule, column)
		}
	}

	// Handle nullable columns
	if column.IsNullable == "YES" && rand.Float32() < 0.1 { // 10% chance of null
		return nil, nil
	}

	// Generate value based on data type
	return dt.generateValueByDataType(column)
}

// applyPrivacyRule applies a privacy rule to generate appropriate fake data
func (dt *DataTransformer) applyPrivacyRule(rule PrivacyRule, column ColumnInfo) (interface{}, error) {
	switch rule.Action {
	case "fake":
		return dt.generateFakeValue(rule.FakerTag, column.DataType)
	case "mask":
		return dt.maskValue(column.DataType)
	case "redact":
		return "[REDACTED]", nil
	case "custom":
		return rule.CustomValue, nil
	default:
		return dt.generateValueByDataType(column)
	}
}

// generateFakeValue generates fake data using the specified faker tag
func (dt *DataTransformer) generateFakeValue(fakerTag, dataType string) (interface{}, error) {
	switch fakerTag {
	case "email":
		return faker.Email(), nil
	case "name":
		return faker.Name(), nil
	case "first_name":
		return faker.FirstName(), nil
	case "last_name":
		return faker.LastName(), nil
	case "username":
		return faker.Username(), nil
	case "phone":
		return faker.Phonenumber(), nil
	case "address":
		return faker.GetRealAddress().Address, nil
	case "city":
		return faker.GetRealAddress().City, nil
	case "company":
		return faker.DomainName(), nil
	case "credit_card":
		return faker.CCNumber(), nil
	case "uuid":
		return uuid.New().String(), nil
	case "password":
		return faker.Password(), nil
	case "url":
		return faker.URL(), nil
	case "ipv4":
		return faker.IPv4(), nil
	case "date":
		return faker.Date(), nil
	case "timestamp":
		return faker.Timestamp(), nil
	default:
		// Fallback to data type generation
		return dt.generateValueByDataType(ColumnInfo{DataType: dataType})
	}
}

// maskValue generates masked values for sensitive data
func (dt *DataTransformer) maskValue(dataType string) (interface{}, error) {
	switch {
	case strings.Contains(dataType, "char") || strings.Contains(dataType, "text"):
		return "****", nil
	case strings.Contains(dataType, "int"):
		return 0, nil
	case strings.Contains(dataType, "decimal") || strings.Contains(dataType, "numeric"):
		return 0.0, nil
	case strings.Contains(dataType, "bool"):
		return false, nil
	default:
		return "****", nil
	}
}

// generateValueByDataType generates a value based on PostgreSQL data type
func (dt *DataTransformer) generateValueByDataType(column ColumnInfo) (interface{}, error) {
	dataType := column.DataType

	// --- 1. Handle Array Types FIRST (before converting to lowercase) ---
	if strings.HasSuffix(dataType, "[]") {
		// Generate a single random word
		word := faker.Word()
		// Format it as a valid PostgreSQL single-element array literal
		// e.g., {"some-random-word"}
		return fmt.Sprintf(`{"%s"}`, word), nil
	}

	// --- 2. Handle Vector Types ---
	if strings.HasPrefix(dataType, "vector") {
		// Use regex to find the dimension, e.g., vector(1536) -> 1536
		re := regexp.MustCompile(`\d+`)
		match := re.FindString(dataType)
		dimension := 3 // Default dimension
		if match != "" {
			if d, err := strconv.Atoi(match); err == nil {
				dimension = d
			}
		}

		vec := make([]string, dimension)
		for i := range vec {
			vec[i] = fmt.Sprintf("%.6f", rand.Float32()*2-1) // Generate a float between -1.0 and 1.0
		}

		// Format as a string literal: "[0.123, -0.456, ...]"
		return fmt.Sprintf("[%s]", strings.Join(vec, ",")), nil
	}

	// --- 3. Handle Tsvector Types ---
	if dataType == "tsvector" {
		words := []string{faker.Word(), faker.Word(), faker.Word()} // Generate 3 random words
		// Format as a valid tsvector literal: 'word1' 'word2' ...
		return fmt.Sprintf("'%s'", strings.Join(words, "' '")), nil
	}

	// Now convert to lowercase for remaining comparisons
	dataType = strings.ToLower(dataType)

	switch {
	// String types
	case strings.Contains(dataType, "character varying"), strings.Contains(dataType, "varchar"):
		length := column.MaxLength
		if length == 0 || length > 100 {
			length = 50 // Default reasonable length
		}
		return faker.Username()[:min(length, len(faker.Username()))], nil

	case strings.Contains(dataType, "character"), strings.Contains(dataType, "char"):
		length := column.MaxLength
		if length == 0 {
			length = 10
		}
		return faker.Username()[:min(length, len(faker.Username()))], nil

	case dataType == "text":
		return faker.Sentence(), nil

	// Integer types
	case dataType == "bigint", dataType == "int8":
		return rand.Int63n(1000000), nil

	case dataType == "integer", dataType == "int", dataType == "int4":
		return rand.Int31n(100000), nil

	case dataType == "smallint", dataType == "int2":
		return rand.Int31n(32767), nil

	// Floating point types
	case dataType == "real", dataType == "float4":
		return rand.Float32() * 1000, nil

	case dataType == "double precision", dataType == "float8":
		return rand.Float64() * 1000, nil

	case strings.Contains(dataType, "numeric"), strings.Contains(dataType, "decimal"):
		return rand.Float64() * 1000, nil

	// Boolean type
	case dataType == "boolean", dataType == "bool":
		return rand.Float32() < 0.5, nil

	// Date and time types
	case dataType == "date":
		return faker.Date(), nil

	case dataType == "time", strings.Contains(dataType, "time"):
		return faker.TimeString(), nil

	case dataType == "timestamp", strings.Contains(dataType, "timestamp"):
		return faker.Timestamp(), nil

	case dataType == "timestamptz", strings.Contains(dataType, "timestamp with time zone"):
		return time.Now().Add(time.Duration(rand.Intn(365*24)) * time.Hour), nil

	// UUID type
	case dataType == "uuid":
		return uuid.New().String(), nil

	// JSON types
	case dataType == "json", dataType == "jsonb":
		return fmt.Sprintf(`{"key": "%s", "random_number": %d}`, faker.Word(), rand.Intn(1000)), nil

	// Binary types
	case dataType == "bytea":
		return faker.Password(), nil

	// Network types
	case dataType == "inet", dataType == "cidr":
		return faker.IPv4(), nil

	case dataType == "macaddr":
		return faker.MacAddress(), nil

	// Default fallback
	default:
		log.Printf("Warning: Unsupported data type '%s' for column, using default string", dataType)
		return faker.Word(), nil
	}
}

// GenerateSyntheticData generates synthetic data for all tables in the schema
func (dt *DataTransformer) GenerateSyntheticData(schema DatabaseSchema) (SyntheticDataResult, error) {
	log.Printf("Starting data generation for %d tables", len(schema.Tables))
	startTime := time.Now()

	result := SyntheticDataResult{
		DatabaseName: schema.DatabaseName + dt.config.DatabaseSuffix,
		GeneratedAt:  time.Now(),
		RowCounts:    make(map[string]int),
	}

	// Determine parallel processing
	parallelTables := dt.config.ParallelTables
	if parallelTables <= 0 {
		parallelTables = 4 // Default to 4 concurrent tables
	}

	semaphore := make(chan struct{}, parallelTables)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var tableData []TableData
	var errors []string

	for _, table := range schema.Tables {
		wg.Add(1)
		go func(tableSchema TableSchema) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			data, err := dt.GenerateTableData(tableSchema, dt.config.RowCount)

			mu.Lock()
			if err != nil {
				errors = append(errors, fmt.Sprintf("Table %s: %v", tableSchema.TableName, err))
				log.Printf("Error generating data for table %s: %v", tableSchema.TableName, err)
			} else {
				tableData = append(tableData, data)
				result.RowCounts[tableSchema.TableName] = len(data.Rows)
				log.Printf("Generated %d rows for table %s", len(data.Rows), tableSchema.TableName)
			}
			mu.Unlock()
		}(table)
	}

	wg.Wait()

	result.Tables = tableData

	if len(errors) > 0 {
		log.Printf("Data generation completed with %d errors", len(errors))
		for _, err := range errors {
			log.Printf("Error: %s", err)
		}
	}

	processingTime := time.Since(startTime)
	log.Printf("Data generation completed in %v. Generated data for %d tables", processingTime, len(tableData))

	return result, nil
}

// Helper function to find minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
