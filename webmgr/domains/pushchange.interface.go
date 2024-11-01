package domains

import "context"

type PushChangeRepository interface {
	ExistsConfiguration(ctx context.Context, cfg *Configure, app *Application) (bool, error)

	InsertNewConfigure(ctx context.Context, cfg *Configure, app *Application) (int64, error)
	UpdateConfigurationSequence(ctx context.Context, cfg *Configure, app *Application, configId int64) (bool, error)
	UpdateConfigure(ctx context.Context, cfg *Configure, app *Application) (bool, error)
}
