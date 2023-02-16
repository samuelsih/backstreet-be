package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"time"
)

const pgTimeout = 30 * time.Second

func ConnectPG(dsn string) (*pgx.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), pgTimeout)
	defer cancel()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}

	return conn, nil
}
