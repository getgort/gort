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
		t.Fatalf("Expected error %q but didn't get one", expected)
	} else if !cogerr.ErrEquals(err, expected) {
		t.Fatalf("Wrong error:\nExpected: %s\nGot: %s\n", expected, err)
	}
}

func expectNoErr(t *testing.T, err error) {
	if err != nil {
		// t.Fatal(err)
		panic(err)
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
