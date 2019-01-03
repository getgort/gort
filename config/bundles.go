package config

import (
	"errors"
	"strings"
)

// CommandEntry wraps a bundle and a command within that bundle.
type CommandEntry struct {
	Bundle  BundleConfig
	Command BundleCommandConfig
}

// FindCommandEntry looks for a command in the configuration. It expects a
// string in the format "bundle:command" or "command"; the latter can return
// multiple values if a similarly-named command is found in multiple bundles.
func FindCommandEntry(name string) ([]CommandEntry, error) {
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

	entries := make([]CommandEntry, 0)

	for _, bundle := range GetBundleConfigs() {
		if bundleName != bundle.Name && bundleName != "*" {
			continue
		}

		for _, command := range bundle.Commands {
			if command.Command == commandName {
				entries = append(entries, CommandEntry{Bundle: bundle, Command: command})
			}
		}
	}

	return entries, nil
}
