package api

import (
	"context"
	"fmt"

	apiv1 "github.com/go-sphere/sphere-bun-layout/api/api/v1"
	"github.com/go-sphere/sphere-bun-layout/api/entpb"
	"github.com/go-sphere/sphere-bun-layout/internal/pkg/conv"
	"github.com/go-sphere/sphere/utils/secure"
	"google.golang.org/protobuf/proto"
)

var _ apiv1.AdminServiceHTTPServer = (*Service)(nil)

func stripPassword(admin *entpb.Admin) *entpb.Admin {
	if admin == nil {
		return nil
	}
	out := proto.Clone(admin).(*entpb.Admin)
	out.Password = ""
	return out
}

func (s *Service) CreateAdmin(ctx context.Context, request *apiv1.CreateAdminRequest) (*apiv1.CreateAdminResponse, error) {
	admin := request.GetAdmin()
	if admin == nil {
		return nil, fmt.Errorf("admin is required")
	}
	hashed, err := secure.CryptPassword(admin.Password)
	if err != nil {
		return nil, err
	}
	toInsert := proto.Clone(admin).(*entpb.Admin)
	toInsert.Id = 0
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
		Admin: stripPassword(&admin),
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
	out := conv.PointerArray(admins)
	for i := range out {
		out[i] = stripPassword(out[i])
	}
	return &apiv1.ListAdminsResponse{
		Admins:    out,
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
		Admin: stripPassword(request.Admin),
	}, nil
}
