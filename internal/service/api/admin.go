package api

import (
	"context"

	apiv1 "github.com/go-sphere/sphere-bun-layout/api/api/v1"
	"github.com/go-sphere/sphere-bun-layout/api/entpb"
	"github.com/go-sphere/sphere-bun-layout/internal/pkg/conv"
	"github.com/go-sphere/sphere/utils/secure"
)

var _ apiv1.AdminServiceHTTPServer = (*Service)(nil)

func (s *Service) CreateAdmin(ctx context.Context, request *apiv1.CreateAdminRequest) (*apiv1.CreateAdminResponse, error) {
	request.Admin.Id = 0
	request.Admin.Password = secure.CryptPassword(request.Admin.Password)
	if _, err := s.db.NewInsert().
		Model(request.Admin).
		Returning("id").
		Exec(ctx); err != nil {
		return nil, err
	}
	return &apiv1.CreateAdminResponse{
		Admin: request.Admin,
	}, nil
}

func (s *Service) DeleteAdmin(ctx context.Context, request *apiv1.DeleteAdminRequest) (*apiv1.DeleteAdminResponse, error) {
	exec, err := s.db.NewDelete().
		Model(&entpb.Admin{Id: request.Id}).
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
	admin := entpb.Admin{Id: request.Id}
	if err := s.db.NewSelect().
		Model(&admin).
		WherePK().
		Scan(ctx); err != nil {
		return nil, err
	}
	return &apiv1.GetAdminResponse{
		Admin: &admin,
	}, nil
}

func (s *Service) ListAdmins(ctx context.Context, request *apiv1.ListAdminsRequest) (*apiv1.ListAdminsResponse, error) {
	var admins []entpb.Admin
	query := s.db.NewSelect().Model(&admins)

	count, err := query.Count(ctx)
	if err != nil {
		return nil, err
	}
	page, size := conv.Page(count, int(request.PageSize))
	err = query.
		OrderExpr("id ASC").
		Limit(size).
		Offset(int(request.Page) * size).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &apiv1.ListAdminsResponse{
		Admins:    conv.PointerArray(admins),
		TotalSize: int64(count),
		TotalPage: int64(page),
	}, nil
}

func (s *Service) UpdateAdmin(ctx context.Context, request *apiv1.UpdateAdminRequest) (*apiv1.UpdateAdminResponse, error) {
	exec, err := s.db.NewUpdate().
		Model(request.Admin).
		Column("name", "email").
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
	return &apiv1.UpdateAdminResponse{
		Admin: request.Admin,
	}, nil
}
