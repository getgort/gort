package client_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/getgort/gort/client"
)

func TestAllowInsecure(t *testing.T) {
	var tests = []struct {
		Name         string
		ProfileEntry client.ProfileEntry
		ExpectErr    bool
	}{
		{
			Name: "allows secure URL",
			ProfileEntry: client.ProfileEntry{
				URLString: "https://example.com",
			},
		},
		{
			Name: "does not allow insecure URL",
			ProfileEntry: client.ProfileEntry{
				URLString: "http://example.com",
			},
			ExpectErr: true,
		},
		{
			Name: "allows insecure URL if allowInsecure==true",
			ProfileEntry: client.ProfileEntry{
				URLString:     "http://example.com",
				AllowInsecure: true,
			},
			ExpectErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			_, err := client.ConnectWithNewProfile(test.ProfileEntry)
			if test.ExpectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
