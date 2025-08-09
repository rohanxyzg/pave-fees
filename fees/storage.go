//go:build !test

package fees

import (
	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("feesdb", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

func getDB() *sqldb.Database {
	return db
}
