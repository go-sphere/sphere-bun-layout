package internal

import (
	"github.com/go-sphere/sphere-bun-layout/internal/biz"
	"github.com/go-sphere/sphere-bun-layout/internal/config"
	"github.com/go-sphere/sphere-bun-layout/internal/pkg"
	"github.com/go-sphere/sphere-bun-layout/internal/server"
	"github.com/go-sphere/sphere-bun-layout/internal/service"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	// Internal application components
	wire.NewSet(
		biz.ProviderSet,
		server.ProviderSet,
		service.ProviderSet,
		config.ProviderSet,
		pkg.ProviderSet,
	),
)
