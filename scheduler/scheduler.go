package scheduler

import (
	"context"
	"github.com/getgort/gort/adapter"
	"github.com/getgort/gort/command"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess"
	"github.com/getgort/gort/telemetry"
	"github.com/go-co-op/gocron"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"time"
)

type CommandScheduler struct {
	cron     *gocron.Scheduler
	Commands chan data.CommandRequest
	da       dataaccess.DataAccess
}

func NewCommandScheduler(da dataaccess.DataAccess) CommandScheduler {
	cron := gocron.NewScheduler(time.Local)
	cron.TagsUnique()
	return CommandScheduler{
		cron:     cron,
		Commands: make(chan data.CommandRequest),
		da:       da,
	}
}

func (cs *CommandScheduler) Start() {
	cs.cron.StartAsync()
}

func (cs *CommandScheduler) schedule(ctx context.Context, command data.ScheduledCommand) error {
	err := cs.da.ScheduleCreate(ctx, &command)
	if err != nil {
		return err
	}
	_, err = cs.cron.Cron(command.Cron).Do(func() { //todo tag with scheduleid
		tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
		ctx, sp := tr.Start(context.Background(), "adapter.handleIncomingEvent")
		defer sp.End()

		req := data.CommandRequest{
			CommandEntry: command.CommandEntry,
			Adapter:      command.Adapter,
			ChannelID:    command.ChannelID,
			Context:      ctx,
			Parameters:   command.Parameters,
			Timestamp:    time.Now(),
			UserID:       command.UserID,
			UserEmail:    command.UserEmail,
			UserName:     command.UserName,
		}

		err := cs.da.RequestBegin(ctx, &req)
		if err != nil {
			sp.RecordError(err)
			sp.SetStatus(codes.Error, "Failed to begin request")
			return
		}

		cs.Commands <- req
	})

	return err
}

func (cs *CommandScheduler) Add(ctx context.Context, cron string, cmdEntry data.CommandEntry, parameters data.CommandParameters, userID, userEmail, userName, adapterName, channelID string) error {

	err := cs.schedule(ctx, data.ScheduledCommand{
		CommandEntry: cmdEntry,
		Adapter:      adapterName,
		ChannelID:    channelID,
		Parameters:   parameters,
		UserID:       userID,
		UserEmail:    userEmail,
		UserName:     userName,
		Cron:         cron,
	})

	return err
}

func (cs *CommandScheduler) AddFromString(ctx context.Context, cron, commandString, userID, userEmail, userName, adapterName, channelID string) error {

	tokens, err := command.Tokenize(commandString)
	if err != nil {
		return err
	}

	cmdEntry, cmdInput, err := adapter.CommandFromTokensByName(ctx, tokens)
	if err != nil {
		return err
	}

	return cs.Add(ctx, cron, *cmdEntry, adapter.ParametersFromCommand(cmdInput), userID, userEmail, userName, adapterName, channelID)
}
