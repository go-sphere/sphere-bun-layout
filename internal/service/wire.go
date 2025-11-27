package service

import (
	"github.com/go-sphere/sphere-bun-layout/internal/service/api"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	api.NewService,
)
