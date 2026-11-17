package api

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/go-sphere/httpx"
	apiv1 "github.com/go-sphere/sphere-bun-layout/api/api/v1"
	"github.com/go-sphere/sphere-bun-layout/api/bunpb"
	"github.com/go-sphere/sphere-bun-layout/internal/pkg/conv"
	"github.com/go-sphere/sphere/utils/secure"
	"google.golang.org/protobuf/proto"
)

var _ apiv1.AdminServiceHTTPServer = (*Service)(nil)

func stripPassword(admin *bunpb.Admin) *bunpb.Admin {
	if admin == nil {
		return nil
	}
	out := proto.Clone(admin).(*bunpb.Admin)
	out.Password = ""
	return out
}

func (s *Service) CreateAdmin(ctx context.Context, request *apiv1.CreateAdminRequest) (*apiv1.CreateAdminResponse, error) {
	admin := request.GetAdmin()
	if admin == nil {
		return nil, httpx.NewBadRequestError("admin is required")
	}
	name := strings.TrimSpace(admin.Name)
	if name == "" {
		return nil, httpx.NewBadRequestError("name is required")
	}
	if len(admin.Password) < 8 || len(admin.Password) > 72 {
		return nil, httpx.NewBadRequestError("password must be between 8 and 72 bytes")
	}
	hashed, err := secure.CryptPassword(admin.Password)
	if err != nil {
		return nil, err
	}
	toInsert := proto.Clone(admin).(*bunpb.Admin)
	toInsert.Id = 0
	toInsert.Name = name
	toInsert.Password = hashed
	if _, err := s.db.NewInsert().
		Model(toInsert).
		Returning("id").
		Exec(ctx); err != nil {
		return nil, err
	}
	return &apiv1.CreateAdminResponse{
		Admin: stripPassword(toInsert),
	}, nil
}

func (s *Service) DeleteAdmin(ctx context.Context, request *apiv1.DeleteAdminRequest) (*apiv1.DeleteAdminResponse, error) {
	exec, err := s.db.NewDelete().
		Model(&bunpb.Admin{Id: request.Id}).
		WherePK().
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	count, err := exec.RowsAffected()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, apiv1.AdminError_ADMIN_ERROR_NOT_FOUND
	}
	return &apiv1.DeleteAdminResponse{}, nil
}

func (s *Service) GetAdmin(ctx context.Context, request *apiv1.GetAdminRequest) (*apiv1.GetAdminResponse, error) {
	admin := &bunpb.Admin{Id: request.Id}
	if err := s.db.NewSelect().
		Model(admin).
		WherePK().
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apiv1.AdminError_ADMIN_ERROR_NOT_FOUND
		}
		return nil, err
	}
	admin.Password = ""
	return &apiv1.GetAdminResponse{
		Admin: admin,
	}, nil
}

func (s *Service) ListAdmins(ctx context.Context, request *apiv1.ListAdminsRequest) (*apiv1.ListAdminsResponse, error) {
	var admins []*bunpb.Admin
	query := s.db.NewSelect().Model(&admins)

	count, err := query.Count(ctx)
	if err != nil {
		return nil, err
	}
	page, size := conv.Page(count, int(request.PageSize))
	if request.Page < 0 || request.Page >= int64(page) {
		return &apiv1.ListAdminsResponse{
			Admins:    []*bunpb.Admin{},
			TotalSize: int64(count),
			TotalPage: int64(page),
		}, nil
	}
	err = query.
		OrderExpr("id ASC").
		Limit(size).
		Offset(int(request.Page) * size).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	for _, a := range admins {
		if a != nil {
			a.Password = ""
		}
	}
	return &apiv1.ListAdminsResponse{
		Admins:    admins,
		TotalSize: int64(count),
		TotalPage: int64(page),
	}, nil
}

func (s *Service) UpdateAdmin(ctx context.Context, request *apiv1.UpdateAdminRequest) (*apiv1.UpdateAdminResponse, error) {
	admin := request.GetAdmin()
	if admin == nil {
		return nil, httpx.NewBadRequestError("admin is required")
	}
	if admin.Id <= 0 {
		return nil, httpx.NewBadRequestError("id must be positive")
	}
	name := strings.TrimSpace(admin.Name)
	if name == "" {
		return nil, httpx.NewBadRequestError("name is required")
	}
	toUpdate := proto.Clone(admin).(*bunpb.Admin)
	toUpdate.Name = name
	cols := []string{"name", "email", "role"}
	if admin.Password != "" {
		if len(admin.Password) < 8 || len(admin.Password) > 72 {
			return nil, httpx.NewBadRequestError("password must be between 8 and 72 bytes")
		}
		hashed, err := secure.CryptPassword(admin.Password)
		if err != nil {
			return nil, err
		}
		toUpdate.Password = hashed
		cols = append(cols, "password")
	}
	exec, err := s.db.NewUpdate().
		Model(toUpdate).
		Column(cols...).
		WherePK().
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	count, err := exec.RowsAffected()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, apiv1.AdminError_ADMIN_ERROR_NOT_FOUND
	}
	updated := &bunpb.Admin{Id: toUpdate.Id}
	if err := s.db.NewSelect().
		Model(updated).
		WherePK().
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apiv1.AdminError_ADMIN_ERROR_NOT_FOUND
		}
		return nil, err
	}
	updated.Password = ""
	return &apiv1.UpdateAdminResponse{
		Admin: updated,
	}, nil
}
