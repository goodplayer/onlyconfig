package domains

import (
	"context"
	"errors"
)

func ApplyConfigureChange(ctx context.Context, repo PushChangeRepository, cfg *Configure, app *Application) error {
	// query existing configure only once since duplicated configuration cannot be inserted
	has, err := repo.ExistsConfiguration(ctx, cfg, app)
	if err != nil {
		return err
	}
	var configId int64
	if !has {
		cfgId, err := repo.InsertNewConfigure(ctx, cfg, app)
		if err != nil {
			return err
		}
		configId = cfgId
	}

	// Retry 10 times to perform update configure CAS
	// This will make sure updated record has the latest sequence
	for i := 0; i < 10; i++ {
		if !has {
			updated, err := repo.UpdateConfigurationSequence(ctx, cfg, app, configId)
			if err != nil {
				return err
			}
			if updated {
				return nil
			}
		} else {
			updated, err := repo.UpdateConfigure(ctx, cfg, app)
			if err != nil {
				return err
			}
			if updated {
				return nil
			}
		}
	}
	return errors.New("max retry exceeded while applying configure")
}
