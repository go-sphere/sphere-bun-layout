package api

import "github.com/uptrace/bun"

// Service implements the demo AdminService on bun. It performs no
// authorization of its own and relies on the server in internal/server/api,
// which only verifies the JWT signature.
type Service struct {
	db *bun.DB
}

func NewService(db *bun.DB) *Service {
	return &Service{
		db: db,
	}
}
