package config

import (
	"errors"
	"strings"

	"github.com/clockworksoul/cog2/data"
)

// FindCommandEntry looks for a command in the configuration. It assumes that
// command character(s) have already been removed, and expects a string in the
// format "bundle:command" or "command"; the latter can return multiple values
// if a similarly-named command is found in multiple bundles
func FindCommandEntry(name string) ([]data.CommandEntry, error) {
	var bundleName string
	var commandName string

	split := strings.Split(name, ":")

	switch len(split) {
	case 1:
		bundleName = "*"
		commandName = split[0]
	case 2:
		bundleName = split[0]
		commandName = split[1]
	default:
		return nil, errors.New("Invalid bundle:comand pair")
	}

	entries := make([]data.CommandEntry, 0)

	for _, bundle := range GetBundleConfigs() {
		if bundleName != bundle.Name && bundleName != "*" {
			continue
		}

		for name, command := range bundle.Commands {

			if name == commandName {

				command.Name = name
				entries = append(entries, data.CommandEntry{
					Bundle:  bundle,
					Command: command,
				})
			}
		}
	}

	return entries, nil
}
