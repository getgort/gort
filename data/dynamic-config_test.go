package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDynamicConfigValidate(t *testing.T) {
	tests := []struct {
		layer string
		err   bool
	}{
		{layer: "bundle", err: false},
		{layer: "room", err: false},
		{layer: "group", err: false},
		{layer: "user", err: false},
		{layer: "Bundle", err: false},
		{layer: "", err: true},
		{layer: "foo", err: true},
	}

	const msg = "layer=%q"

	for _, test := range tests {
		layer := ConfigurationLayer(test.layer)

		if test.err {
			assert.Error(t, layer.Validate(), msg, test.layer)
		} else {
			assert.NoError(t, layer.Validate(), msg, test.layer)
		}
	}
}
