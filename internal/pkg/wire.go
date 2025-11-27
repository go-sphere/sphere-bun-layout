package pkg

import (
	"github.com/go-sphere/sphere-bun-layout/internal/pkg/database"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	wire.NewSet(
		database.NewDatabase,
		database.NewDbConnection,
	),
)
