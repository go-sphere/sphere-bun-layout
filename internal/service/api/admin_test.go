package api

import (
	"net/http"
	"testing"

	"github.com/go-sphere/httpx"
	apiv1 "github.com/go-sphere/sphere-bun-layout/api/api/v1"
	"github.com/go-sphere/sphere-bun-layout/api/bunpb"
	"github.com/go-sphere/sphere-bun-layout/internal/pkg/database"
	"github.com/uptrace/bun"
)

func newTestDB(t *testing.T) *bun.DB {
	t.Helper()
	sqlDB, err := database.NewDbConnection(database.Config{Location: "file::memory:"})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	// One connection: every new connection to :memory: is a fresh database.
	sqlDB.SetMaxOpenConns(1)
	db, err := database.NewDatabase(sqlDB)
	if err != nil {
		t.Fatalf("wrap database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.NewCreateTable().Model((*bunpb.Admin)(nil)).Exec(t.Context()); err != nil {
		t.Fatalf("create admin table: %v", err)
	}
	return db
}

func TestGetAdminMissingIsNotFound(t *testing.T) {
	_, err := NewService(newTestDB(t)).GetAdmin(t.Context(), &apiv1.GetAdminRequest{Id: 999})
	if err == nil {
		t.Fatal("GetAdmin(missing) error = nil")
	}
	if _, status, _ := httpx.ParseError(err); status != http.StatusNotFound {
		t.Fatalf("GetAdmin(missing) status = %d, want %d (err = %v)", status, http.StatusNotFound, err)
	}
}
