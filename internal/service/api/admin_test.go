package api

import (
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-sphere/httpx"
	apiv1 "github.com/go-sphere/sphere-bun-layout/api/api/v1"
	"github.com/go-sphere/sphere-bun-layout/api/bunpb"
	"github.com/go-sphere/sphere-bun-layout/internal/pkg/database"
	"github.com/go-sphere/sphere/utils/secure"
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

func TestAdminServiceCRUDValidationAndPagination(t *testing.T) {
	ctx := t.Context()
	db := newTestDB(t)
	svc := NewService(db)

	for _, tc := range []struct {
		name string
		req  *apiv1.CreateAdminRequest
	}{
		{name: "nil admin", req: &apiv1.CreateAdminRequest{}},
		{name: "empty name", req: &apiv1.CreateAdminRequest{Admin: &bunpb.Admin{Name: "   ", Password: "ValidPassword#1"}}},
		{name: "short password", req: &apiv1.CreateAdminRequest{Admin: &bunpb.Admin{Name: "alice", Password: "short"}}},
		{name: "over 72 byte password", req: &apiv1.CreateAdminRequest{Admin: &bunpb.Admin{Name: "alice", Password: strings.Repeat("a", 73)}}},
	} {
		_, err := svc.CreateAdmin(ctx, tc.req)
		if _, status, _ := httpx.ParseError(err); status != http.StatusBadRequest {
			t.Errorf("CreateAdmin(%s) status = %d (err=%v), want 400", tc.name, status, err)
		}
	}

	created, err := svc.CreateAdmin(ctx, &apiv1.CreateAdminRequest{
		Admin: &bunpb.Admin{
			Name:     " alice ",
			Email:    "alice@example.com",
			Role:     "viewer",
			Password: "InitialPassword#1",
		},
	})
	if err != nil {
		t.Fatalf("CreateAdmin() error = %v", err)
	}
	if created.Admin.Id <= 0 || created.Admin.Name != "alice" || created.Admin.Password != "" {
		t.Fatalf("unexpected created admin: %+v", created.Admin)
	}

	for _, tc := range []struct {
		name string
		req  *apiv1.UpdateAdminRequest
	}{
		{name: "nil admin", req: &apiv1.UpdateAdminRequest{}},
		{name: "zero id", req: &apiv1.UpdateAdminRequest{Admin: &bunpb.Admin{Id: 0, Name: "alice"}}},
		{name: "empty name", req: &apiv1.UpdateAdminRequest{Admin: &bunpb.Admin{Id: created.Admin.Id, Name: "  "}}},
		{name: "short password", req: &apiv1.UpdateAdminRequest{Admin: &bunpb.Admin{Id: created.Admin.Id, Name: "alice", Password: "short"}}},
		{name: "over 72 byte password", req: &apiv1.UpdateAdminRequest{Admin: &bunpb.Admin{Id: created.Admin.Id, Name: "alice", Password: strings.Repeat("a", 73)}}},
	} {
		_, err := svc.UpdateAdmin(ctx, tc.req)
		if _, status, _ := httpx.ParseError(err); status != http.StatusBadRequest {
			t.Errorf("UpdateAdmin(%s) status = %d (err=%v), want 400", tc.name, status, err)
		}
	}

	updated, err := svc.UpdateAdmin(ctx, &apiv1.UpdateAdminRequest{
		Admin: &bunpb.Admin{
			Id:       created.Admin.Id,
			Name:     "alice-updated",
			Email:    "alice2@example.com",
			Role:     "admin",
			Password: "NewPassword#2",
		},
	})
	if err != nil {
		t.Fatalf("UpdateAdmin() error = %v", err)
	}
	if updated.Admin.Name != "alice-updated" || updated.Admin.Email != "alice2@example.com" || updated.Admin.Role != "admin" || updated.Admin.Password != "" {
		t.Fatalf("unexpected updated admin: %+v", updated.Admin)
	}

	persisted := &bunpb.Admin{Id: created.Admin.Id}
	if err := db.NewSelect().Model(persisted).WherePK().Scan(ctx); err != nil {
		t.Fatalf("reload updated admin: %v", err)
	}
	if !secure.IsPasswordMatch("NewPassword#2", persisted.Password) {
		t.Fatal("UpdateAdmin did not persist new password hash")
	}

	// Empty password on update keeps the existing hash.
	if _, err := svc.UpdateAdmin(ctx, &apiv1.UpdateAdminRequest{
		Admin: &bunpb.Admin{Id: created.Admin.Id, Name: "alice-keep", Email: "alice2@example.com", Role: "editor"},
	}); err != nil {
		t.Fatalf("UpdateAdmin(empty password) error = %v", err)
	}
	if err := db.NewSelect().Model(persisted).WherePK().Scan(ctx); err != nil {
		t.Fatalf("reload admin: %v", err)
	}
	if persisted.Role != "editor" || !secure.IsPasswordMatch("NewPassword#2", persisted.Password) {
		t.Fatalf("unexpected persisted admin after role-only update: role=%q", persisted.Role)
	}

	listed, err := svc.ListAdmins(ctx, &apiv1.ListAdminsRequest{Page: 0, PageSize: 9999})
	if err != nil {
		t.Fatalf("ListAdmins() error = %v", err)
	}
	if len(listed.Admins) != 1 || listed.Admins[0].Password != "" {
		t.Fatalf("unexpected ListAdmins response: %+v", listed.Admins)
	}

	outOfRange, err := svc.ListAdmins(ctx, &apiv1.ListAdminsRequest{Page: math.MaxInt64, PageSize: 10})
	if err != nil {
		t.Fatalf("ListAdmins(out of range) error = %v", err)
	}
	if len(outOfRange.Admins) != 0 || outOfRange.TotalSize != 1 {
		t.Fatalf("unexpected out-of-range ListAdmins response: %+v", outOfRange)
	}
}

func TestNewDbConnectionCreatesSQLiteDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "db")
	sqlDB, err := database.NewDbConnection(database.Config{Location: filepath.Join(dir, "app.sqlite3")})
	if err != nil {
		t.Fatalf("NewDbConnection() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("expected dir %s to be created, err=%v", dir, err)
	}
}
