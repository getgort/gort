package memory

import (
	"testing"

	gorterr "github.com/clockworksoul/gort/errors"
	"github.com/stretchr/testify/assert"
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

func testInitialize(t *testing.T) {
	da = NewInMemoryDataAccess()

	err := da.Initialize()
	assert.NoError(t, err)
}

func TestMain(t *testing.T) {
	t.Run("testInitialize", testInitialize)

	t.Run("testUserAccess", testUserAccess)
	t.Run("testGroupAccess", testGroupAccess)
	t.Run("testTokenAccess", testTokenAccess)
	t.Run("testBundleAccess", testBundleAccess)
}
