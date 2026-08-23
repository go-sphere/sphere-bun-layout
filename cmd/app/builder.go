package main

import (
	"github.com/go-sphere/sphere-bun-layout/internal/biz/task"
	"github.com/go-sphere/sphere-bun-layout/internal/server/api"
	"github.com/go-sphere/sphere/core/boot"
	coretask "github.com/go-sphere/sphere/core/task"
)

func newApplication(api *api.Web, dbInit *task.DbInit) *boot.Application {
	// dbInit starts first (create tables) and stops last (close DB) so HTTP
	// requests are not served against a closed connection.
	return boot.NewStagedApplication(
		[]coretask.Task{dbInit},
		[]coretask.Task{api},
	)
}
