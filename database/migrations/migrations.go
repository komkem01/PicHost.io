package migrations

import (
	"embed"

	"github.com/uptrace/bun/migrate"
)

var Migrations = migrate.NewMigrations()

//go:embed *
var mFile embed.FS

func init() {
	if err := Migrations.Discover(mFile); err != nil {
		panic(err)
	}
}
