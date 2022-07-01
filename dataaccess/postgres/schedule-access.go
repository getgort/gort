package postgres

import (
	"context"
	"fmt"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/telemetry"
	"go.opentelemetry.io/otel"
)

//TODO(grebneerg): Assign unique IDs and put in database here.

func (da PostgresDataAccess) ScheduleCreate(ctx context.Context, command *data.ScheduledCommand) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "postgres.ScheduleCreate")
	defer sp.End()

	if command.ScheduleID != 0 {
		return fmt.Errorf("Schedule id already set.")
	}

	command.ScheduleID++

	return nil
}

func (da PostgresDataAccess) ScheduleDelete(ctx context.Context, command data.ScheduledCommand) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "postgres.ScheduleDelete")
	defer sp.End()

	if command.ScheduleID == 0 {
		return fmt.Errorf("Schedule id not set.")
	}

	return nil
}
