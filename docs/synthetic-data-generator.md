# VapusAI Synthetic Data Generator

A production-ready ETL pipeline for generating realistic synthetic test data from PostgreSQL databases. Creates isolated test databases with privacy-compliant fake data.

## Features

- ✅ **Production Safe** - Read-only on source, creates isolated test databases
- ✅ **Complete PostgreSQL Support** - All data types including arrays, vectors, JSON, UUIDs
- ✅ **Privacy Rules** - Configurable data masking for sensitive columns
- ✅ **Zero Hardcoding** - Uses environment variables for credentials
- ✅ **Concurrent Processing** - Parallel table processing for performance

## Quick Start

### 1. Build
```bash
cd app/cli
make build
```

### 2. Set Environment Variables
```bash
export DB_HOST="your-database-host"
export DB_USER="postgres"
export DB_PASSWORD="your-password"
export DB_NAME="your-database-name"
```

### 3. Generate Data
```bash
# Test connection
./vapusctl synth-data --config ../../examples/synth-data-config.yaml --validate-only

# Generate synthetic data
./vapusctl synth-data --config ../../examples/synth-data-config.yaml --suffix _synthetic_test --rows 100
```

Your synthetic database will be created as `{DB_NAME}_synthetic_test`.

## Configuration

The example config at `examples/synth-data-config.yaml` supports environment variables:

```yaml
# Source database
source_db:
  host: "${DB_HOST:-localhost}"
  port: ${DB_PORT:-5432}
  username: "${DB_USER:-postgres}"
  password: "${DB_PASSWORD:-your_password}"
  database: "${DB_NAME:-vapusaidev}"
  sslmode: "${DB_SSL_MODE:-disable}"

# Target database (will be created)
target_db:
  host: "${DB_HOST:-localhost}"
  port: ${DB_PORT:-5432}
  username: "${DB_USER:-postgres}"
  password: "${DB_PASSWORD:-your_password}"
  database: "${DB_NAME:-vapusaidev}"
  sslmode: "${DB_SSL_MODE:-disable}"

# Generation settings
database_suffix: "_synthetic_test"
row_count: 100
parallel_tables: 3

# Privacy rules for sensitive data
privacy_rules:
  - column_pattern: ".*email.*"
    faker_tag: "email"
  - column_pattern: ".*password.*"
    faker_tag: "password"
  - column_pattern: ".*phone.*"
    faker_tag: "phone"
```

## Command Reference

| Flag | Description | Example |
|------|-------------|---------|
| `--config, -c` | Configuration file path (required) | `--config config.yaml` |
| `--rows, -r` | Number of rows per table | `--rows 1000` |
| `--suffix, -s` | Database suffix | `--suffix "_test"` |
| `--parallel, -p` | Parallel table processing | `--parallel 4` |
| `--dry-run` | Preview schema only | |
| `--validate-only` | Test connections only | |
| `--generate-config` | Create config template | `--generate-config config.yaml` |

## Common Usage

```bash
# Generate config template
./vapusctl synth-data --generate-config my-config.yaml

# Test connections
./vapusctl synth-data --config ../../examples/synth-data-config.yaml --validate-only

# Preview schema
./vapusctl synth-data --config ../../examples/synth-data-config.yaml --dry-run

# Generate small test dataset
./vapusctl synth-data --config ../../examples/synth-data-config.yaml --suffix _test --rows 10

# Generate full dataset
./vapusctl synth-data --config ../../examples/synth-data-config.yaml --suffix _synthetic_test --rows 10000
```

## Privacy Rules

Configure automatic masking of sensitive data using column pattern matching:

```yaml
privacy_rules:
  - column_pattern: ".*email.*"    # Matches any column containing "email"
    faker_tag: "email"             # Generates fake email addresses
  - column_pattern: ".*name.*"
    faker_tag: "name"
  - column_pattern: ".*phone.*"
    faker_tag: "phone"
  - column_pattern: ".*address.*"
    faker_tag: "address"
```

Supported faker tags: `email`, `name`, `phone`, `password`, `company`, `address`, `city`, `uuid`

## Supported Data Types

✅ All PostgreSQL types including:
- Numeric: `integer`, `bigint`, `decimal`, `serial`
- Text: `varchar`, `text`, `char`
- Date/Time: `timestamp`, `date`, `time`
- Special: `json`, `jsonb`, `uuid`, `boolean`
- Arrays: `text[]`, `integer[]`, etc.
- Extensions: `vector` (pgvector), `tsvector`

## Troubleshooting

**Connection Error**: Check host, port, credentials in config  
**Permission Error**: Ensure user has `CREATE DATABASE` privileges  
**Memory Issues**: Reduce `row_count` or `parallel_tables`

```bash
# Debug mode
./vapusctl synth-data --config ../../examples/synth-data-config.yaml --validate-only

# Small test run
./vapusctl synth-data --config ../../examples/synth-data-config.yaml --suffix _test --rows 5
```

## Output

Successful generation creates:
- New database: `{source_database}_synthetic_test`
- Result file: `synthetic_data_result_YYYYMMDD_HHMMSS.json`
- Console summary with table and row counts

---

**Ready to generate synthetic data?** Use the Quick Start guide above! 
