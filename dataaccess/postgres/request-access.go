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
	"fmt"
	"strings"

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess/errs"
	gerr "github.com/getgort/gort/errors"
	"github.com/getgort/gort/telemetry"
	"go.opentelemetry.io/otel"
)

func (da PostgresDataAccess) RequestBegin(ctx context.Context, req *data.CommandRequest) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.RequestBegin")
	defer sp.End()

	if req.RequestID != 0 {
		return fmt.Errorf("command request ID already set")
	}

	db, err := da.connect(ctx, DatabaseGort)
	if err != nil {
		return err
	}
	defer db.Close()

	const query = `INSERT INTO commands (bundle_name, bundle_version, command_name,
		command_executable, command_parameters, adapter, user_id,
		user_email, channel_id, gort_user_name, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING request_id;`

	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx,
		req.Bundle.Name,
		req.Bundle.Version,
		req.Command.Name,
		encodeStringSlice(req.Command.Executable),
		strings.Join(req.Parameters, " "),
		req.Adapter,
		req.UserID,
		req.UserEmail,
		req.ChannelID,
		req.UserName,
		req.Timestamp).Scan(&req.RequestID)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

func (da PostgresDataAccess) RequestError(ctx context.Context, req data.CommandRequest, err error) error {
	response := data.CommandResponse{
		Command: req,
		Error:   err,
		Status:  1,
	}

	return da.RequestClose(ctx, response)
}

func (da PostgresDataAccess) RequestUpdate(ctx context.Context, req data.CommandRequest) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.RequestUpdate")
	defer sp.End()

	if req.RequestID == 0 {
		return fmt.Errorf("command request ID unset")
	}

	db, err := da.connect(ctx, DatabaseGort)
	if err != nil {
		return err
	}
	defer db.Close()

	const query = `UPDATE commands
		SET bundle_name=$1, bundle_version=$2, command_name=$3,
			command_executable=$4, command_parameters=$5, adapter=$6, user_id=$7,
			user_email=$8, channel_id=$9, gort_user_name=$10
		WHERE request_id=$11;`

	_, err = db.ExecContext(ctx, query,
		req.Bundle.Name,
		req.Bundle.Version,
		req.Command.Name,
		encodeStringSlice(req.Command.Executable),
		strings.Join(req.Parameters, " "),
		req.Adapter,
		req.UserID,
		req.UserEmail,
		req.ChannelID,
		req.UserName,
		req.RequestID)
	if err != nil {
		err = gerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

func (da PostgresDataAccess) RequestClose(ctx context.Context, res data.CommandResponse) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "postgres.RequestClose")
	defer sp.End()

	if res.Command.RequestID == 0 {
		return fmt.Errorf("command request ID unset")
	}

	db, err := da.connect(ctx, DatabaseGort)
	if err != nil {
		return err
	}
	defer db.Close()

	const query = `UPDATE commands
		SET bundle_name=$1, bundle_version=$2, command_name=$3,
			command_executable=$4, command_parameters=$5, adapter=$6, user_id=$7,
			user_email=$8, channel_id=$9, gort_user_name=$10, timestamp=$11,
			duration=$12, result_status=$13, result_error=$14
		WHERE request_id=$15;`

	errMsg := ""
	if res.Error != nil {
		errMsg = res.Error.Error()
	}

	_, err = db.ExecContext(ctx, query,
		res.Command.Bundle.Name,
		res.Command.Bundle.Version,
		res.Command.Command.Name,
		encodeStringSlice(res.Command.Command.Executable),
		strings.Join(res.Command.Parameters, " "),
		res.Command.Adapter,
		res.Command.UserID,
		res.Command.UserEmail,
		res.Command.ChannelID,
		res.Command.UserName,
		res.Command.Timestamp,
		res.Duration.Milliseconds(),
		res.Status,
		errMsg,
		res.Command.RequestID)
	if err != nil {
		err = gerr.Wrap(errs.ErrDataAccess, err)
	}

	return err
}

func (da PostgresDataAccess) createCommandsTable(ctx context.Context, db *sql.DB) error {
	createCommandsQuery := `CREATE TABLE commands(
		request_id          BIGSERIAL,
		timestamp           TIMESTAMP WITH TIME ZONE,
		duration            INT,
		bundle_name			TEXT NOT NULL,
		bundle_version		TEXT NOT NULL,
		command_name		TEXT NOT NULL,
		command_executable	TEXT NOT NULL,
	    command_parameters  TEXT NOT NULL,
		adapter		        TEXT NOT NULL,
		user_id		        TEXT NOT NULL,
		user_email		    TEXT NOT NULL,
		channel_id		    TEXT NOT NULL,
		gort_user_name      TEXT NOT NULL,
		result_status		INT,
		result_error        TEXT
	);`

	_, err := db.ExecContext(ctx, createCommandsQuery)
	if err != nil {
		return gerr.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}
