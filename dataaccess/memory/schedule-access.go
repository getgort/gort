package memory

import (
	"context"
	"fmt"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/telemetry"
	"go.opentelemetry.io/otel"
)

func (da *InMemoryDataAccess) ScheduleCreate(ctx context.Context, command *data.ScheduledCommand) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "memory.ScheduleCreate")
	defer sp.End()

	if command.ScheduleID != 0 {
		return fmt.Errorf("Schedule id already set.")
	}

	command.ScheduleID++

	return nil
}

func (da *InMemoryDataAccess) ScheduleDelete(ctx context.Context, command data.ScheduledCommand) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "memory.ScheduleDelete")
	defer sp.End()

	if command.ScheduleID == 0 {
		return fmt.Errorf("Schedule id not set.")
	}

	return nil
}
