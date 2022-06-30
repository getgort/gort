package scheduler

import (
	"context"
	"fmt"
	"github.com/getgort/gort/adapter"
	"github.com/getgort/gort/command"
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
	cs.da.ScheduleCreate(ctx, &command)
	fmt.Println("scheduling")
	_, err := cs.cron.Cron(command.Cron).Do(func() { //todo tag with scheduleid
		fmt.Println("ccron time")
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

		cs.da.RequestBegin(ctx, &req)

		cs.Commands <- req
	})

	return err
}

func (cs *CommandScheduler) Add(ctx context.Context, cron, commandString, userID, userEmail, UserName, adapterName, channelID string) error {
	tokens, err := command.Tokenize(commandString)
	if err != nil {
		return err
	}

	fmt.Println(tokens)

	cmdEntry, cmdInput, err := adapter.CommandFromTokensByName(ctx, tokens)
	if err != nil {
		return err
	}

	//fmt.Println(cmdEntry)
	//fmt.Println(cmdInput)

	err = cs.schedule(ctx, data.ScheduledCommand{
		CommandEntry: *cmdEntry,
		Adapter:      adapterName,
		ChannelID:    channelID,
		Parameters:   adapter.ParametersFromCommand(cmdInput),
		UserID:       userID,
		UserEmail:    userEmail,
		UserName:     UserName,
		Cron:         cron,
	})
	if err != nil {
		return err
	}

	return nil
}
