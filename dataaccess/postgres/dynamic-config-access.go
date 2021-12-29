/*
 * Copyright 2021 The Gort Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package postgres

import (
	"context"
	"database/sql"

	"go.opentelemetry.io/otel"

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess/errs"
	gerr "github.com/getgort/gort/errors"
	"github.com/getgort/gort/telemetry"
)

func (da PostgresDataAccess) DynamicConfigurationCreate(ctx context.Context, dc data.DynamicConfiguration) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.DynamicConfigurationCreate")
	defer sp.End()

	if err := validate(dc.Layer, dc.Bundle, dc.Owner, dc.Key); err != nil {
		return err
	}

	if exists, err := da.DynamicConfigurationExists(ctx, dc.Layer, dc.Bundle, dc.Owner, dc.Key); err != nil {
		return err
	} else if exists {
		return errs.ErrConfigExists
	}

	db, err := da.connect(ctx, DatabaseGort)
	if err != nil {
		return err
	}
	defer db.Close()

	query := `INSERT INTO configs
		(bundle_name, layer, owner, key, secret, value)
		VALUES ($1, $2, $3, $4, $5, $6);`
	_, err = db.ExecContext(ctx, query, dc.Bundle, dc.Layer, dc.Owner, dc.Key, dc.Secret, dc.Value)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

func (da PostgresDataAccess) DynamicConfigurationDelete(ctx context.Context, layer data.ConfigurationLayer, bundle, owner, key string) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.DynamicConfigurationDelete")
	defer sp.End()

	if err := validate(layer, bundle, owner, key); err != nil {
		return err
	}

	if exists, err := da.DynamicConfigurationExists(ctx, layer, bundle, owner, key); err != nil {
		return err
	} else if !exists {
		return errs.ErrNoSuchConfig
	}

	db, err := da.connect(ctx, DatabaseGort)
	if err != nil {
		return err
	}
	defer db.Close()

	query := "DELETE FROM configs WHERE bundle_name=$1 AND layer=$2 AND owner=$3 AND key=$4;"
	_, err = db.ExecContext(ctx, query, bundle, layer, owner, key)
	if err != nil {
		err = gerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

func (da PostgresDataAccess) DynamicConfigurationExists(ctx context.Context, layer data.ConfigurationLayer, bundle, owner, key string) (bool, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.DynamicConfigurationExists")
	defer sp.End()

	if err := validate(layer, bundle, owner, key); err != nil {
		return false, err
	}

	db, err := da.connect(ctx, DatabaseGort)
	if err != nil {
		return false, err
	}
	defer db.Close()

	query := "SELECT EXISTS(SELECT 1 FROM configs WHERE bundle_name=$1 AND layer=$2 AND owner=$3 AND key=$4)"
	exists := false

	err = db.QueryRowContext(ctx, query, bundle, layer, owner, key).Scan(&exists)
	if err != nil {
		return false, gerr.Wrap(errs.ErrNoSuchGroup, err)
	}

	return exists, nil
}

func (da PostgresDataAccess) DynamicConfigurationGet(ctx context.Context, layer data.ConfigurationLayer, bundle, owner, key string) (data.DynamicConfiguration, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.DynamicConfigurationGet")
	defer sp.End()

	if err := validate(layer, bundle, owner, key); err != nil {
		return data.DynamicConfiguration{}, err
	}

	if exists, err := da.DynamicConfigurationExists(ctx, layer, bundle, owner, key); err != nil {
		return data.DynamicConfiguration{}, err
	} else if !exists {
		return data.DynamicConfiguration{}, errs.ErrNoSuchConfig
	}

	db, err := da.connect(ctx, DatabaseGort)
	if err != nil {
		return data.DynamicConfiguration{}, err
	}
	defer db.Close()

	query := `SELECT bundle_name, layer, owner, key, value, secret
		FROM configs
		WHERE bundle_name=$1 AND layer=$2 AND owner=$3 AND key=$4`
	dc := data.DynamicConfiguration{}

	err = db.QueryRowContext(ctx, query, bundle, layer, owner, key).
		Scan(&dc.Bundle, &dc.Layer, &dc.Owner, &dc.Key, &dc.Value, &dc.Secret)

	if err == sql.ErrNoRows {
		return dc, errs.ErrNoSuchGroup
	} else if err != nil {
		return dc, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return dc, nil
}

func (da PostgresDataAccess) DynamicConfigurationList(ctx context.Context, layer data.ConfigurationLayer, bundle, owner, key string) ([]data.DynamicConfiguration, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.DynamicConfigurationList")
	defer sp.End()

	if bundle == "" {
		return nil, errs.ErrEmptyConfigBundle
	}
	if layer == "" {
		layer = "%"
	}
	if owner == "" {
		owner = "%"
	}
	if key == "" {
		key = "%"
	}

	var dcs = make([]data.DynamicConfiguration, 0)

	db, err := da.connect(ctx, DatabaseGort)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `SELECT bundle_name, layer, owner, key, value, secret
		FROM configs
		WHERE bundle_name LIKE $1 AND layer LIKE $2 AND owner LIKE $3 AND key LIKE $4`

	rows, err := db.QueryContext(ctx, query, bundle, layer, owner, key)
	if err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}

	for rows.Next() {
		dc := data.DynamicConfiguration{}

		err = rows.Scan(&dc.Bundle, &dc.Layer, &dc.Owner, &dc.Key, &dc.Value, &dc.Secret)
		if err != nil {
			return nil, gerr.Wrap(errs.ErrNoSuchGroup, err)
		}

		dcs = append(dcs, dc)
	}

	if rows.Err(); err != nil {
		return nil, gerr.Wrap(errs.ErrDataAccess, err)
	}

	return dcs, nil
}

func validate(layer data.ConfigurationLayer, bundle, owner, key string) error {
	switch {
	case bundle == "":
		return errs.ErrEmptyConfigBundle
	case layer == "":
		return errs.ErrEmptyConfigLayer
	case layer.Validate() != nil:
		return layer.Validate()
	case owner == "":
		return errs.ErrEmptyConfigOwner
	case key == "":
		return errs.ErrEmptyConfigKey
	}

	return nil
}
