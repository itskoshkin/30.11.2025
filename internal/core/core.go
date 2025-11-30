package core

import (
	"go.uber.org/fx"

	"link-availability-checker/internal/api"
	"link-availability-checker/internal/api/controllers"
	"link-availability-checker/internal/config"
	"link-availability-checker/internal/logger"
	"link-availability-checker/internal/services"
	"link-availability-checker/internal/storage"
	"link-availability-checker/pkg/filestore"
)

func Load() *fx.App {
	return fx.New(
		config.MuteFxLog(),
		fx.Invoke(
			config.LoadConfig,
			logger.SetupLogging,
		),
		fx.Provide(
			filestore.NewFileStorer,
			storage.NewLinkStorage,
			services.NewAvailabilityService,
			services.NewLinkService,
			controllers.NewLinkController,
			controllers.NewSystemController,
			api.NewEngine,
		),
		fx.Invoke(
			api.RegisterRoutes,
			api.Run,
		),
	)
}
