package server

import (
	"github.com/go-sphere/sphere-bun-layout/internal/server/api"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	api.NewWebServer,
)
