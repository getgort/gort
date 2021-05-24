package memory

import (
	"testing"

	gorterr "github.com/clockworksoul/gort/errors"
)

var (
	da *InMemoryDataAccess
)

func expectErr(t *testing.T, err error, expected error) {
	if err == nil {
		t.Error("Expected an error")
	} else if !gorterr.Is(err, expected) {
		t.Errorf("Wrong error: Expected: %q Got: %q\n", expected.Error(), err.Error())
	}
}

func expectNoErr(t *testing.T, err error) {
	if err != nil {
		t.Error("Expected no error. Got:", err.Error())
	}
}

func TestDataAccessInit(t *testing.T) {
	da = NewInMemoryDataAccess()

	err := da.Initialize()
	expectNoErr(t, err)
}
