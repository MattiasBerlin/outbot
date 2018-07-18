package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // db driver
)

const (
	user     = "outbot"
	password = "outbot"
	dbname   = "outbot"
)

// New database connection.
func New() (*sql.DB, error) {
	connStr := fmt.Sprintf("user=%v password=%v dbname=%v", user, password, dbname)
	return sql.Open("postgres", connStr)
}
