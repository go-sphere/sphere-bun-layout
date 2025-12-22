package api

import (
	"context"

	"github.com/go-sphere/httpx"
	apiv1 "github.com/go-sphere/sphere-bun-layout/api/api/v1"
	"github.com/go-sphere/sphere-bun-layout/internal/pkg/httpsrv"
	"github.com/go-sphere/sphere-bun-layout/internal/service/api"
)

type Web struct {
	config  *Config
	server  httpx.Engine
	service *api.Service
}

func NewWebServer(conf *Config, service *api.Service) *Web {
	return &Web{
		config:  conf,
		server:  httpsrv.NewFiberServer("api", conf.HTTP.Address),
		service: service,
	}
}

func (w *Web) Identifier() string {
	return "api"
}

func (w *Web) Start(ctx context.Context) error {
	route := w.server.Group("/")
	apiv1.RegisterAdminServiceHTTPServer(route, w.service)
	return w.server.Start()
}

func (w *Web) Stop(ctx context.Context) error {
	return w.server.Stop(ctx)
}
