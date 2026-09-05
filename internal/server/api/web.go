package api

import (
	"context"

	"github.com/go-sphere/httpx"
	apiv1 "github.com/go-sphere/sphere-bun-layout/api/api/v1"
	"github.com/go-sphere/sphere-bun-layout/internal/pkg/httpsrv"
	"github.com/go-sphere/sphere-bun-layout/internal/service/api"
	"github.com/go-sphere/sphere/server/auth/jwtauth"
	"github.com/go-sphere/sphere/server/middleware/auth"
)

type Web struct {
	config  Config
	server  httpx.Engine
	service *api.Service
}

func NewWebServer(conf Config, service *api.Service) *Web {
	return &Web{
		config:  conf,
		server:  httpsrv.NewGinServer("api", conf.HTTP.Address),
		service: service,
	}
}

func (w *Web) Identifier() string {
	return "api"
}

func (w *Web) Start(ctx context.Context) error {
	if err := httpsrv.UseCORS(w.server, w.config.HTTP.Cors); err != nil {
		return err
	}
	jwtAuthorizer := jwtauth.NewJwtAuth[jwtauth.RBACClaims[int64]](w.config.JWT)
	authMiddleware := auth.NewAuthMiddleware[int64, jwtauth.RBACClaims[int64]](
		jwtAuthorizer,
		auth.WithHeaderLoader(auth.AuthorizationHeader),
		auth.WithPrefixTransform(auth.AuthorizationPrefixBearer),
		auth.WithAbortOnError(true),
	)
	route := w.server.Group("/", authMiddleware)
	apiv1.RegisterAdminServiceHTTPServer(route, w.service)
	return w.server.Start()
}

func (w *Web) Stop(ctx context.Context) error {
	return w.server.Stop(ctx)
}
