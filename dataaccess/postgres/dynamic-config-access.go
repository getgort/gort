package postgres

import (
	"context"
	"fmt"

	"github.com/getgort/gort/data"
)

func (da PostgresDataAccess) DynamicConfigurationCreate(_ context.Context, config data.DynamicConfiguration) error {
	return fmt.Errorf("not implemented")
}

func (da PostgresDataAccess) DynamicConfigurationDelete(_ context.Context, layer data.ConfigurationLayer, bundle, owner, key string) error {
	return fmt.Errorf("not implemented")
}

func (da PostgresDataAccess) DynamicConfigurationExists(_ context.Context, layer data.ConfigurationLayer, bundle, owner, key string) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (da PostgresDataAccess) DynamicConfigurationGet(_ context.Context, layer data.ConfigurationLayer, bundle, owner, key string) (data.DynamicConfiguration, error) {
	return data.DynamicConfiguration{}, fmt.Errorf("not implemented")
}

func (da PostgresDataAccess) DynamicConfigurationList(_ context.Context, layer data.ConfigurationLayer, bundle, owner, key string) ([]data.DynamicConfiguration, error) {
	return nil, fmt.Errorf("not implemented")
}
