package main

import (
	"github.com/go-sphere/sphere-bun-layout/internal/biz/task"
	"github.com/go-sphere/sphere-bun-layout/internal/server/api"
	"github.com/go-sphere/sphere/core/boot"
)

func newApplication(api *api.Web, dbInit *task.DbInit) *boot.Application {
	return boot.NewApplication(
		dbInit,
		api,
	)
}
