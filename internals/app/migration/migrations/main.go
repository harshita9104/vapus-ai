package migrations

import (
	"embed"

	"github.com/uptrace/bun/migrate"
)

var Migrations = migrate.NewMigrations()

//go:embed *.sql
var sqlMigrations embed.FS

func init() {
	// Discover Go-based migrations
	if err := Migrations.DiscoverCaller(); err != nil {
		panic(err)
	}

	// Discover SQL-based migrations
	if err := Migrations.Discover(sqlMigrations); err != nil {
		panic(err)
	}
}
