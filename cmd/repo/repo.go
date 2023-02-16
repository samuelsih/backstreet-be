package repo

import (
	"backstreetlinkv2/cmd/model"
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5"
)

var (
	UniqueErr   = errors.New("link already taken")
	NotFoundErr = errors.New("not found")
	InternalErr = errors.New("internal error")
)

type PGRepo struct {
	db *pgx.Conn
}

func (p *PGRepo) Insert(ctx context.Context, key string, dataSource any) error {
	const query = `INSERT INTO source (key_source, attrs) VALUES ($1, $2)`

	cmd, err := p.db.Exec(ctx, query, key, dataSource)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return InternalErr
	}

	return nil
}

func (p *PGRepo) Get(ctx context.Context, key string) (model.ShortenResponse, error) {
	const query = ` SELECT attrs FROM source WHERE key_source = $1`

	var resp model.ShortenResponse

	err := p.db.QueryRow(ctx, query, key).Scan(&resp)

	if err != nil {
		if err == pgx.ErrNoRows || err == sql.ErrNoRows {
			return resp, NotFoundErr
		}

		return resp, InternalErr
	}

	return resp, nil
}

func NewPGRepo(db *pgx.Conn) *PGRepo {
	return &PGRepo{db: db}
}
