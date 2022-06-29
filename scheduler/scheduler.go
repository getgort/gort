package scheduler

import (
	"context"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess"
	"github.com/getgort/gort/telemetry"
	"github.com/go-co-op/gocron"
	"go.opentelemetry.io/otel"
	"time"
)

type CommandScheduler struct {
	cron     *gocron.Scheduler
	Commands chan data.CommandRequest
	da       dataaccess.DataAccess
}

func NewCommandScheduler(da dataaccess.DataAccess) CommandScheduler {
	return CommandScheduler{
		cron:     gocron.NewScheduler(time.Local),
		Commands: make(chan data.CommandRequest),
		da:       da,
	}
}

func (cs *CommandScheduler) Start() {
	cs.cron.StartAsync()
}

func (cs *CommandScheduler) Add(command data.ScheduledCommand) {
	//TODO check permissions
	cs.cron.Cron(command.Cron).Do(func() {
		tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
		ctx, sp := tr.Start(context.Background(), "adapter.handleIncomingEvent")
		defer sp.End()

		req := data.CommandRequest{
			CommandEntry: command.CommandEntry,
			Adapter:      command.Adapter,
			ChannelID:    command.ChannelID,
			Context:      ctx,
			Parameters:   command.Parameters,
			RequestID:    0,
			Timestamp:    time.Now(),
			UserID:       command.UserID,
			UserEmail:    command.UserEmail,
			UserName:     command.UserName,
		}

		cs.da.RequestBegin(ctx, &req)

		cs.Commands <- req
	})
}
