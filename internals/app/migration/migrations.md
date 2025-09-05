# VapusAI Database Migration System - Testing Guide

## What This Does
This tool helps you change your database safely. It tracks what changes you make so you can undo them if needed.

## Quick Setup

### Step 1: Build the Tool
```bash
cd workspace/vapus-ai/internals/app/migration/cmd
go build -o vapus-migrate simple-main.go
```
**What this creates**: A binary file called `vapus-migrate` that you can run directly.

### Step 2: Test Connection
```bash
# Set your database details and test connection replace DB_PASSWORD with actual password
export DB_HOST="35.244.50.2"
export DB_PASSWORD="your_password"
export DB_NAME="vapusaidev"
./vapus-migrate status
```
**What you should see**: Connection successful + list of migrations

## Complete Test - Proven Working Commands

### Test 1: Check Current Status replace DB_PASSWORD with actual password
```bash

cd /home/harshita/workspace/vapus-ai/internals/app/migration/cmd && export DB_HOST="35.244.50.2" && export DB_PASSWORD="your_password" && export DB_NAME="vapusaidev" && echo "=== TESTING CONNECTION ===" && ./vapus-migrate status
```

### Test 2: Create New Migration
```bash
echo "=== CREATING TEST MIGRATION ===" && ./vapus-migrate create-sql test_column_addition
```

### Test 3: Check for New Pending Migration
```bash
echo "=== CHECKING STATUS - SHOULD SHOW NEW PENDING MIGRATION ===" && ./vapus-migrate status
```

### Test 4: Apply Migration (Makes Database Changes)
```bash
echo "=== APPLYING MIGRATION - WILL ADD test_notes_column TO users TABLE ===" && ./vapus-migrate migrate
```

### Test 5: Check Status After Migration
```bash
echo "=== CHECKING STATUS AFTER MIGRATION ===" && ./vapus-migrate status
```

### Test 6: Test Rollback (Undo Changes)
```bash
echo "=== TESTING ROLLBACK - WILL REMOVE test_notes_column ===" && ./vapus-migrate rollback
```

### Test 7: Check Status After Rollback
```bash
echo "=== CHECKING STATUS AFTER ROLLBACK ===" && ./vapus-migrate status
```

## What Each Command Does

| Command | What It Does |
|---------|-------------|
| `status` | Shows what migrations are applied/pending |
| `create-sql <name>` | Creates new migration files for you to edit |
| `migrate` | Applies pending migrations to database |
| `rollback` | Undoes the last set of migrations |
| `init` | Sets up migration tracking (run once) |
| `unlock` | Fixes stuck migrations |

## Files Created During Setup

When you build the tool:
- **`vapus-migrate`** - The main binary file you run (created by `go build`)
- **Migration files** - `.up.sql` and `.down.sql` files (created by `create-sql` command)

**Note**: The `vapus-migrate` binary is what you use to run all migration commands.

## For Your Database (Replace with your details)
```bash
# Set these first before any command:
export DB_HOST="your-database-ip"
export DB_PASSWORD="your-password" 
export DB_NAME="your-database-name"
```

## Testing Success Signs
- **Connection works**: You see "Connected to database" message
- **Migrations work**: Status shows migrations are applied/pending correctly  
- **Database changes**: You can see actual changes in your database
- **Rollback works**: Changes get undone when you rollback

## Common Commands

### Check What's Happening
```bash
./vapus-migrate status
```

### Apply All Pending Changes
```bash
./vapus-migrate migrate
```

### Undo Last Changes
```bash
./vapus-migrate rollback
```

### Create New Migration
```bash
./vapus-migrate create-sql my_new_change
```

### Fix Stuck Migrations
```bash
./vapus-migrate unlock
```

## How Migration Files Work

When you create a migration, you get 2 files:
- **`.up.sql`** - What changes to make (add column, create table, etc.)
- **`.down.sql`** - How to undo those changes (remove column, drop table, etc.)

Edit these files with your actual SQL before running `migrate`.

## Example: Adding a Column

1. **Create migration**: `./vapus-migrate create-sql add_user_phone`
2. **Edit the `.up.sql` file**:
   ```sql
   ALTER TABLE users ADD COLUMN phone VARCHAR(20);
   ```
3. **Edit the `.down.sql` file**:
   ```sql
   ALTER TABLE users DROP COLUMN phone;
   ```
4. **Apply it**: `./vapus-migrate migrate`
5. **Check it worked**: `./vapus-migrate status`

## Troubleshooting

**Can't connect?** Check your DB_PASSWORD and DB_HOST are correct.

**Migration stuck?** Run `./vapus-migrate unlock`

**Want to see what's pending?** Run `./vapus-migrate status`

**Made a mistake?** Run `./vapus-migrate rollback` to undo

## Use


1. **Build the tool first**: `go build -o vapus-migrate simple-main.go`
2. **Test migrations on a copy** of your database first
3. **Back up your database** before running migrations  


## Important Notes About vapus-migrate Binary

- The `vapus-migrate` file is a **compiled binary** created by the `go build` command
- This binary contains all the migration functionality
- You can copy this binary to any server and run migrations
- Always rebuild the binary when you update the source code
- The binary is specific to your operating system (Linux/Windows/Mac)



