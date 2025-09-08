# VapusAI Go-Based Migration System

## Setup

### Build Tool
```bash
cd /home/harshita/workspace/vapus-ai/internals/app/migration/cmd
go build -o simple-main simple-main.go
```

### Database Credentials (.env file)
```bash
DB_HOST=your_host
DB_PASSWORD=your_password
DB_NAME=your_database
DB_USER=postgres
DB_PORT=5432
DB_SSLMODE=disable
```

## Commands

### Create Migration
```bash
./simple-main creation <migration_name>
```
Creates `timestamp_migration_name.go` file ready to run.

### Check Status
```bash
./simple-main status
```
Shows applied/pending migrations and sequence.

### Run Migrations
```bash
./simple-main :run
```
Applies all pending migrations top-to-bottom.

### Rollback
```bash
./simple-main revert
```
Undoes last migration group.

### Other Commands
```bash
./simple-main init      # Initialize migration table
./simple-main unlock    # Fix stuck migrations
```

## Migration File Format

Generated files contain:
- UP function: Creates tables/changes
- DOWN function: Rollback logic  
- Error handling: IF EXISTS/IF NOT EXISTS
- Sample data: Ready to use

## Database Verification

### Check tables created:
```sql
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public';
```

### Check migration tracking:
```sql
SELECT name, group_id, migrated_at FROM bun_migrations 
ORDER BY migrated_at DESC;
```

### Check table data:
```sql
SELECT * FROM your_table_name_models;
```

## Key Features

✅ **Single file**: Both UP and DOWN in one Go file  
✅ **Unique timestamps**: Automatic file naming  
✅ **Ready to run**: No editing required  
✅ **Database tracking**: All migrations tracked  
✅ **Rollback support**: Full undo functionality  
✅ **Environment config**: Uses .env file  

## Workflow

1. Create migration: `./simple-main creation <name>`
2. Run migrations: `./simple-main :run`
3. Check status: `./simple-main status`
4. Rollback if needed: `./simple-main revert`  




