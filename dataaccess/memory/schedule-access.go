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

package memory

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/telemetry"
	"go.opentelemetry.io/otel"
)

func (da *InMemoryDataAccess) ScheduleCreate(ctx context.Context, command *data.ScheduledCommand) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "memory.ScheduleCreate")
	defer sp.End()

	if command.ScheduleID != 0 {
		return fmt.Errorf("schedule id already set")
	}

	command.ScheduleID = atomic.AddInt64(&da.schedules.id, 1)

	da.schedules.schedules[command.ScheduleID] = command

	return nil
}

func (da *InMemoryDataAccess) ScheduleDelete(ctx context.Context, command data.ScheduledCommand) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "memory.ScheduleDelete")
	defer sp.End()

	if command.ScheduleID == 0 {
		return fmt.Errorf("schedule id not set")
	}

	_, found := da.schedules.schedules[command.ScheduleID]

	if !found {
		return fmt.Errorf("schedule %d not found", command.ScheduleID)
	}

	delete(da.schedules.schedules, command.ScheduleID)

	return nil
}

func (da *InMemoryDataAccess) SchedulesGet(ctx context.Context) ([]data.ScheduledCommand, error) {
	s := make([]data.ScheduledCommand, 0)

	for _, c := range da.schedules.schedules {
		s = append(s, *c)
	}

	return s, nil
}
