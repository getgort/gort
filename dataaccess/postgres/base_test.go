package postgres

import (
	"testing"

	"github.com/clockworksoul/cog2/data"
	cogerr "github.com/clockworksoul/cog2/errors"
)

var (
	da PostgresDataAccess
)

func expectErr(t *testing.T, err error, expected error) {
	if err == nil {
		t.Error("Expected an error")
	} else if !cogerr.ErrEquals(err, expected) {
		t.Errorf("Wrong error:\nExpected: %s\nGot: %s\n", expected.Error(), err.Error())
	}
}

func expectNoErr(t *testing.T, err error) {
	if err != nil {
		t.Error("Expected no error. Got:", err.Error())
	}
}

func TestDataAccessInit(t *testing.T) {
	configs := data.DatabaseConfigs{
		Host:       "localhost",
		Password:   "password",
		Port:       5432,
		SSLEnabled: false,
		User:       "cog",
	}

	da = NewPostgresDataAccess(configs)

	err := da.Initialize()
	expectNoErr(t, err)
}
