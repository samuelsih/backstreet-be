package repo

import (
	"backstreetlinkv2/api/helper"
	"backstreetlinkv2/api/model"
	"context"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"errors"
)

var (
	ErrUnique         = errors.New("link already taken")
	ErrNotFound       = errors.New("not found")
	ErrNoRowsAffected = errors.New("no rows affected")
)

const (
	UniqueConstraint   = 1062
	CantProcessRequest = "can't process your request"
)

type MYSQLRepo struct {
	db *sql.DB
}

func (p *MYSQLRepo) Insert(ctx context.Context, key string, dataSource any) error {
	const op = helper.Op("repo.MYSQLRepo.Insert")
	const query = `INSERT INTO sources (key_source, attrs) VALUES (?, ?)`

	cmd, err := p.db.ExecContext(ctx, query, key, dataSource)
	if err != nil {
		var mysqlErr *mysql.MySQLError

		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == UniqueConstraint {
				return helper.E(op, helper.KindBadRequest, ErrUnique, ErrUnique.Error())
			}
		}

		return helper.E(op, helper.KindUnexpected, err, CantProcessRequest)
	}

	rowsAffected, err := cmd.RowsAffected()
	if err != nil {
		helper.E(op, helper.KindUnexpected, err, CantProcessRequest)
	}

	if rowsAffected == 0 {
		return helper.E(op, helper.KindUnexpected, ErrNoRowsAffected, CantProcessRequest)
	}

	return nil
}

func (p *MYSQLRepo) Get(ctx context.Context, key string) (model.ShortenResponse, error) {
	const op = helper.Op("repo.MYSQLRepo.Get")
	const query = `SELECT attrs FROM sources WHERE key_source = ?`

	var resp model.ShortenResponse

	err := p.db.QueryRowContext(ctx, query, key).Scan(&resp)

	if err != nil {
		if err == sql.ErrNoRows {
			return resp, helper.E(op, helper.KindNotFound, ErrNotFound, ErrNotFound.Error())
		}

		return resp, helper.E(op, helper.KindUnexpected, err, CantProcessRequest)
	}

	return resp, nil
}

func NewMYSQLRepo(db *sql.DB) *MYSQLRepo {
	return &MYSQLRepo{db: db}
}
