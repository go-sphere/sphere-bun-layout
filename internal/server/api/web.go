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

// Web is the demo API server. It only checks that a request carries a JWT
// signed with Config.JWT: there is no login endpoint, no token issuance and no
// role check, so any token holder can create and delete admins. Before real
// use, add an endpoint that issues tokens and protect AdminService with a
// permission middleware (see sphere-layout's dash server for both).
type Web struct {
	config  Config
	server  httpx.Engine
	service *api.Service
}

func NewWebServer(conf Config, service *api.Service) *Web {
	return &Web{
		config:  conf,
		server:  httpsrv.NewServer("api", conf.HTTP.Address, conf.HTTP.Options),
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
