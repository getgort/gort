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

package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/getgort/gort/auth"
	"github.com/getgort/gort/command"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess"
	"github.com/getgort/gort/retrieval"
	"github.com/getgort/gort/telemetry"

	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

var (
	cron        *gocron.Scheduler
	commandsOut = make(chan data.CommandRequest, 100)
)

// StartScheduler starts the scheduler running and returns a channel of
// data.CommandRequest according to registered schedules.
func StartScheduler() chan data.CommandRequest {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(context.Background(), "scheduler.StartScheduler")
	defer sp.End()

	if cron == nil {
		cron = gocron.NewScheduler(time.Local)
		cron.TagsUnique()
	}

	if !cron.IsRunning() {
		schedules, err := GetSchedules(ctx)
		if err != nil {
			sp.RecordError(err)
			log.Errorf("Failed to get schedules: %s", err)
		} else {
			for _, s := range schedules {
				err := schedule(s)
				if err != nil {
					sp.RecordError(err)
					log.WithField("ScheduleID", s.ScheduleID).Warnf("Failed to schedule command")
				}
			}
		}
		cron.StartAsync()
	}

	return commandsOut
}

// StopScheduler stops the scheduler from running. Blocks to wait for any
// currently running jobs to complete.
func StopScheduler() {
	if cron != nil && cron.IsRunning() {
		cron.Stop()
		cron.Clear()
	}
}

func schedule(cmd data.ScheduledCommand) error {
	if cmd.ScheduleID == 0 {
		return fmt.Errorf("scheduled command not initialized")
	}

	_, err := cron.Cron(cmd.Cron).Tag(fmt.Sprintf("%d", cmd.ScheduleID)).Do(func() {
		tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
		ctx, sp := tr.Start(context.Background(), "scheduler.Schedule.cronFunc")
		defer sp.End()

		req := data.CommandRequest{
			CommandEntry: cmd.CommandEntry,
			Adapter:      cmd.Adapter,
			ChannelID:    cmd.ChannelID,
			Context:      ctx,
			Parameters:   retrieval.ParametersFromCommand(cmd.Command),
			Timestamp:    time.Now(),
			UserID:       cmd.UserID,
			UserEmail:    cmd.UserEmail,
			UserName:     cmd.UserName,
		}

		da, err := dataaccess.Get()
		if err != nil {
			sp.RecordError(err)
			sp.SetStatus(codes.Error, "failed to get DataAccess")
			return
		}
		err = da.RequestBegin(ctx, &req)
		if err != nil {
			sp.RecordError(err)
			sp.SetStatus(codes.Error, "failed to begin request")
			return
		}

		commandsOut <- req
	})

	return err
}

// Schedule registers a data.ScheduledCommand with the scheduler so it will be
// requested appropriately.
func Schedule(ctx context.Context, cmd data.ScheduledCommand) (int64, error) {
	err := auth.CheckPermissions(ctx, cmd.UserName, cmd.Command, cmd.CommandEntry)
	if err != nil {
		return 0, err
	}
	da, err := dataaccess.Get()
	if err != nil {
		return 0, err
	}
	err = da.ScheduleCreate(ctx, &cmd)
	if err != nil {
		return 0, err
	}

	return cmd.ScheduleID, schedule(cmd)
}

// ScheduleFromString schedules a command using its string representation.
func ScheduleFromString(ctx context.Context, commandString string, etc data.ScheduledCommand) (int64, error) {
	tokens, err := command.Tokenize(commandString)
	if err != nil {
		return 0, err
	}

	cmdEntry, cmdInput, err := retrieval.CommandFromTokensByName(ctx, tokens)
	if err != nil {
		return 0, err
	}

	cmdInput.Original = commandString

	etc.CommandEntry = *cmdEntry
	etc.Command = cmdInput

	return Schedule(ctx, etc)
}

// Cancel cancels the scheduled command with the given id. Even in the event of
// an error, the command may still be cancelled.
func Cancel(ctx context.Context, scheduleID int64) error {
	da, err := dataaccess.Get()
	if err != nil {
		return err
	}

	err = da.ScheduleDelete(ctx, scheduleID)
	if err != nil {
		return err
	}

	err = cron.RemoveByTag(fmt.Sprintf("%d", scheduleID))
	if err != nil {
		return err
	}

	return nil
}

func GetSchedules(ctx context.Context) ([]data.ScheduledCommand, error) {
	da, err := dataaccess.Get()
	if err != nil {
		return []data.ScheduledCommand{}, err
	}

	schedules, err := da.SchedulesGet(ctx)
	for i := range schedules {
		tokens, err := command.Tokenize(schedules[i].Command.Original)
		if err != nil {
			return []data.ScheduledCommand{}, err
		}

		cmdEntry, cmdInput, err := retrieval.CommandFromTokensByName(ctx, tokens)
		if err != nil {
			return []data.ScheduledCommand{}, err
		}

		cmdInput.Original = schedules[i].Command.Original
		schedules[i].CommandEntry = *cmdEntry
		schedules[i].Command = cmdInput
	}

	return schedules, nil
}
