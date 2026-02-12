package task

import (
	"context"

	"github.com/go-sphere/sphere-bun-layout/api/entpb"
	"github.com/uptrace/bun"
)

type DbInit struct {
	db *bun.DB
}

func NewDbInit(db *bun.DB) *DbInit {
	return &DbInit{
		db: db,
	}
}

func (d DbInit) Identifier() string {
	return "dbinit"
}

func (d DbInit) Start(ctx context.Context) error {
	_, err := d.db.NewCreateTable().IfNotExists().Model(&entpb.Admin{}).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (d DbInit) Stop(ctx context.Context) error {
	return d.db.Close()
}
