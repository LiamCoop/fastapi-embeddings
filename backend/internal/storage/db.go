package storage

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	ErrEmptyDSN = errors.New("database dsn is required")
)

var (
	dbOnce sync.Once
	dbConn *sql.DB
	dbErr  error
)

// OpenDB initializes and returns a singleton database connection pool.
func OpenDB(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, ErrEmptyDSN
	}

	dbOnce.Do(func() {
		dbConn, dbErr = sql.Open("pgx", dsn)
		if dbErr != nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		dbErr = dbConn.PingContext(ctx)
	})

	return dbConn, dbErr
}
