package biz

import (
	"github.com/go-sphere/sphere-bun-layout/internal/biz/task"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	task.NewDbInit,
)
