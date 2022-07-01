package scheduler

import (
	"context"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess/memory"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const OneSecond = 1*time.Second + 50*time.Millisecond

func TestCommandSchedulerFull(t *testing.T) {
	cs := NewCommandScheduler(memory.NewInMemoryDataAccess())
	cs.Start()
	cs.Add(context.Background(), "@every 1s", data.CommandEntry{
		Bundle:  data.Bundle{},
		Command: data.BundleCommand{},
	}, make(data.CommandParameters, 0), "id", "email", "name", "test adapter", "channel")

	select {
	case <-time.After(OneSecond):
		t.Fail()
	case req := <-cs.Commands:
		assert.NotEqual(t, 0, req.RequestID)
		// success
	}
}
