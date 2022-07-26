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
	"fmt"

	"github.com/getgort/gort/command"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess/errs"
	gerrs "github.com/getgort/gort/errors"
	"github.com/getgort/gort/telemetry"

	"go.opentelemetry.io/otel"
)

func (da PostgresDataAccess) ScheduleCreate(ctx context.Context, command *data.ScheduledCommand) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "postgres.ScheduleCreate")
	defer sp.End()

	if command.ScheduleID != 0 {
		return fmt.Errorf("schedule ID already set")
	}

	conn, err := da.connect(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	query := `INSERT INTO schedules
		(adapter, channel_id, user_id, username, user_email, cron, command)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING schedule_id;`

	err = conn.QueryRowContext(
		ctx, query, command.Adapter, command.ChannelID, command.UserID,
		command.UserName, command.UserEmail, command.Cron,
		command.Command.String(),
	).Scan(&command.ScheduleID)

	if err != nil {
		return gerrs.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

func (da PostgresDataAccess) ScheduleDelete(ctx context.Context, scheduleID int64) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "postgres.ScheduleDelete")
	defer sp.End()

	if scheduleID == 0 {
		return fmt.Errorf("schedule id not set")
	}

	conn, err := da.connect(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	query := `DELETE FROM schedules WHERE schedule_id=$1;`
	_, err = conn.ExecContext(ctx, query, scheduleID)
	if err != nil {
		return gerrs.Wrap(errs.ErrDataAccess, err)
	}

	return nil
}

func (da PostgresDataAccess) SchedulesGet(ctx context.Context) ([]data.ScheduledCommand, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "postgres.SchedulesGet")
	defer sp.End()

	schedules := make([]data.ScheduledCommand, 0)

	conn, err := da.connect(ctx)
	if err != nil {
		return schedules, err
	}
	defer conn.Close()

	query := `SELECT schedule_id, adapter, channel_id, user_id, username, user_email, cron, command FROM schedules;`
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return schedules, gerrs.Wrap(errs.ErrDataAccess, err)
	}
	defer rows.Close()

	for rows.Next() {
		var s data.ScheduledCommand
		var c string
		err := rows.Scan(&s.ScheduleID, &s.Adapter, &s.ChannelID, &s.UserID, &s.UserName, &s.UserEmail, &s.Cron, &c)
		if err != nil {
			return nil, err
		}

		s.Command, err = command.TokenizeAndParse(c)
		schedules = append(schedules, s)
	}

	return schedules, nil
}
