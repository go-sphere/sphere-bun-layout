package database

import (
	"database/sql"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

type Config struct {
	Location string `json:"location" yaml:"location"`
}

func NewDbConnection(conf Config) (*sql.DB, error) {
	if err := ensureSQLiteDir(conf.Location); err != nil {
		return nil, err
	}
	return sql.Open(sqliteshim.ShimName, conf.Location)
}

func NewDatabase(db *sql.DB) (*bun.DB, error) {
	return bun.NewDB(db, sqlitedialect.New()), nil
}

func ensureSQLiteDir(dsn string) error {
	pathPart, queryPart, _ := strings.Cut(dsn, "?")
	if query, err := url.ParseQuery(queryPart); err == nil && query.Get("mode") == "memory" {
		return nil
	}
	pathPart = strings.TrimPrefix(pathPart, "file:")
	if strings.HasPrefix(pathPart, "//localhost/") {
		pathPart = strings.TrimPrefix(pathPart, "//localhost")
	}
	if pathPart == "" || pathPart == ":memory:" {
		return nil
	}
	cleanPath := filepath.Clean(pathPart)
	dir := filepath.Dir(cleanPath)
	if dir == "" || dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}
