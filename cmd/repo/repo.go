package repo

import (
	"backstreetlinkv2/cmd/helper"
	"backstreetlinkv2/cmd/model"
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	UniqueErr         = errors.New("link already taken")
	NotFoundErr       = errors.New("not found")
	NoRowsAffectedErr = errors.New("no rows affected")
)

const (
	UniqueConstraint   = "23505"
	CantProcessRequest = "can't process your request"
)

type PGRepo struct {
	db *pgx.Conn
}

func (p *PGRepo) Insert(ctx context.Context, key string, dataSource any) error {
	const op = helper.Op("repo.PGRepo.Insert")
	const query = `INSERT INTO sources (key_source, attrs) VALUES ($1, $2)`

	cmd, err := p.db.Exec(ctx, query, key, dataSource)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == UniqueConstraint {
				return helper.E(op, helper.KindBadRequest, UniqueErr, UniqueErr.Error())
			}
		}

		return helper.E(op, helper.KindUnexpected, err, CantProcessRequest)
	}

	if cmd.RowsAffected() == 0 {
		return helper.E(op, helper.KindUnexpected, NoRowsAffectedErr, CantProcessRequest)
	}

	return nil
}

func (p *PGRepo) Get(ctx context.Context, key string) (model.ShortenResponse, error) {
	const op = helper.Op("repo.PGRepo.Get")
	const query = `SELECT attrs FROM sources WHERE key_source = $1`

	var resp model.ShortenResponse

	err := p.db.QueryRow(ctx, query, key).Scan(&resp)

	if err != nil {
		if err == pgx.ErrNoRows || err == sql.ErrNoRows {
			return resp, helper.E(op, helper.KindNotFound, NotFoundErr, NotFoundErr.Error())
		}

		return resp, helper.E(op, helper.KindUnexpected, err, CantProcessRequest)
	}

	return resp, nil
}

func NewPGRepo(db *pgx.Conn) *PGRepo {
	return &PGRepo{db: db}
}
