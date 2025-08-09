package fees

import (
	"sync"
	"encore.dev/storage/sqldb"
)

var (
	db   *sqldb.Database
	dbOnce sync.Once
)

func getDB() *sqldb.Database {
	dbOnce.Do(func() {
		db = sqldb.NewDatabase("feesdb", sqldb.DatabaseConfig{
			Migrations: "./migrations",
		})
	})
	return db
}
