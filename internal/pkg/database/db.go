package database

import (
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

type Config struct {
	Location string
}

func NewDbConnection(conf Config) (*sql.DB, error) {
	return sql.Open(sqliteshim.ShimName, conf.Location)
}

func NewDatabase(db *sql.DB) (*bun.DB, error) {
	return bun.NewDB(db, sqlitedialect.New()), nil
}
