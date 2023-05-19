package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const mySqlTimeout = 30 * time.Second

func ConnectMySQL(uri string) (*sql.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mySqlTimeout)
	defer cancel()

	db, err := sql.Open("mysql", uri)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
