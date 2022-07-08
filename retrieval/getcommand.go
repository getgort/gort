/*
 * Copyright 2021 The Gort Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package retrieval

import (
	"context"
	"errors"
	"strings"

	"github.com/getgort/gort/bundles"
	"github.com/getgort/gort/command"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/dataaccess"
	gerrs "github.com/getgort/gort/errors"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrNoSuchCommand is returned by GetCommandEntry if a request command
	// isn't found.
	ErrNoSuchCommand = errors.New("no such bundle")

	// ErrMultipleCommands is returned by GetCommandEntry when the same command
	// shortcut matches commands in two or more bundles.
	ErrMultipleCommands = errors.New("multiple commands match that pattern")
)

// CommandFromTokens defines a function that attempts to identify a command from a slice of tokens.
// It returns both a data.CommandEntry defining the command, and a command.Command that re-defines the input
// as appropriate to the command that was found.
type CommandFromTokens func(ctx context.Context, tokens []string) (*data.CommandEntry, command.Command, error)

// CommandFromTokensByName implements CommandFromTokens.
// It checks if a command can be identified from the given tokens by the command name.
func CommandFromTokensByName(ctx context.Context, tokens []string) (*data.CommandEntry, command.Command, error) {
	// Build a temporary Command value using default tokenization rules. We'll
	// use this to load the CommandEntry for the relevant command (as defined
	// in a command bundle), which contains the command's parsing rules that
	// we'll use for a final, formal Parse to get the final Command version.
	cmdInput, err := command.Parse(tokens)
	if err != nil {
		return nil, command.Command{}, err
	}

	cmdEntry, err := GetCommandEntry(ctx, cmdInput.Bundle, cmdInput.Command)
	if err != nil {
		return nil, command.Command{}, err
	}

	// Now that we have a command entry, we can re-create the complete Command value.
	tokens[0] = cmdEntry.Bundle.Name + ":" + cmdEntry.Command.Name

	// TODO Set parse options based on the CommandEntry settings.
	cmdInput, err = command.Parse(tokens)
	if err != nil {
		return nil, command.Command{}, err
	}

	return &cmdEntry, cmdInput, nil
}

// CommandFromTokensByTrigger implements CommandFromTokens.
// It checks if a command can be identified from the given tokens by a trigger pattern.
func CommandFromTokensByTrigger(ctx context.Context, tokens []string) (*data.CommandEntry, command.Command, error) {
	cmdEntry, err := GetCommandEntryByTrigger(ctx, tokens)
	if err != nil && gerrs.Is(err, ErrNoSuchCommand) {
		return nil, command.Command{}, nil
	}
	if err != nil {
		return nil, command.Command{}, err
	}

	// TODO Set parse options based on the CommandEntry settings.
	cmdInput, err := command.Parse(
		append(
			[]string{cmdEntry.Bundle.Name + ":" + cmdEntry.Command.Name},
			tokens...,
		),
	)
	if err != nil {
		return nil, command.Command{}, err
	}
	return &cmdEntry, cmdInput, err
}

// CommandFromTokensByNameOrTrigger implements CommandFromTokens.
// It first checks if a command can be identified from the given tokens by name,
// if this is unsuccessful because the command does not exist, it will attempt to
// identify the command from a trigger.
func CommandFromTokensByNameOrTrigger(ctx context.Context, tokens []string) (*data.CommandEntry, command.Command, error) {
	cmdEntry, cmdInput, err := CommandFromTokensByName(ctx, tokens)
	if err == nil {
		return cmdEntry, cmdInput, nil
	}
	if err != nil && !errors.Is(err, ErrNoSuchCommand) {
		return nil, command.Command{}, err
	}
	return CommandFromTokensByTrigger(ctx, tokens)
}

// ParametersFromCommand converts parameters from a command.Command into
// a string slice.
func ParametersFromCommand(cmd command.Command) []string {
	var out []string
	for _, p := range cmd.Parameters {
		out = append(out, p.String())
	}
	return out
}

// GetCommandEntry accepts a tokenized parameter slice and returns any
// associated data.CommandEntry instances. If the number of matching
// commands is > 1, an error is returned.
func GetCommandEntry(ctx context.Context, bundleName, commandName string) (data.CommandEntry, error) {
	finders, err := allCommandEntryFinders()
	if err != nil {
		return data.CommandEntry{}, err
	}

	entries, err := findAllEntries(ctx, bundleName, commandName, finders...)
	if err != nil {
		return data.CommandEntry{}, err
	}

	if len(entries) == 0 {
		return data.CommandEntry{}, ErrNoSuchCommand
	}

	if len(entries) > 1 {
		cmd := commandName
		if bundleName != "" {
			cmd = bundleName + ":" + commandName
		}

		log.WithField("requested", cmd).
			WithField("bundle0", entries[0].Bundle.Name).
			WithField("command0", entries[0].Command.Name).
			WithField("bundle1", entries[1].Bundle.Name).
			WithField("command1", entries[1].Command.Name).
			Warn("Multiple commands found")

		return data.CommandEntry{}, ErrMultipleCommands
	}

	return entries[0], nil
}

// GetCommandEntryByTrigger accepts a tokenized parameter slice and returns any
// associated data.CommandEntry instances. If the number of matching
// commands is > 1, an error is returned.
func GetCommandEntryByTrigger(ctx context.Context, tokens []string) (data.CommandEntry, error) {
	finders, err := allCommandEntryFinders()
	if err != nil {
		return data.CommandEntry{}, err
	}

	entries, err := findAllEntriesByTrigger(ctx, tokens, finders...)
	if err != nil {
		return data.CommandEntry{}, err
	}

	if len(entries) == 0 {
		return data.CommandEntry{}, ErrNoSuchCommand
	}

	if len(entries) > 1 {
		log.WithField("requested", strings.Join(tokens, " ")).
			WithField("bundle0", entries[0].Bundle.Name).
			WithField("command0", entries[0].Command.Name).
			WithField("bundle1", entries[1].Bundle.Name).
			WithField("command1", entries[1].Command.Name).
			Warn("Multiple commands found")

		return data.CommandEntry{}, ErrMultipleCommands
	}

	return entries[0], nil
}

func allCommandEntryFinders() ([]bundles.CommandEntryFinder, error) {
	finders := make([]bundles.CommandEntryFinder, 0)

	// Get the DAL CommandEntryFinder
	dal, err := dataaccess.Get()
	if err != nil {
		return nil, err
	}

	finders = append(finders, dal)

	return finders, nil
}

func findAllEntries(ctx context.Context, bundleName, commandName string, finder ...bundles.CommandEntryFinder) ([]data.CommandEntry, error) {
	entries := make([]data.CommandEntry, 0)

	for _, f := range finder {
		e, err := f.FindCommandEntry(ctx, bundleName, commandName)
		if err != nil {
			return nil, err
		}

		entries = append(entries, e...)
	}

	return entries, nil
}

func findAllEntriesByTrigger(ctx context.Context, tokens []string, finder ...bundles.CommandEntryFinder) ([]data.CommandEntry, error) {
	entries := make([]data.CommandEntry, 0)

	for _, f := range finder {
		e, err := f.FindCommandEntryByTrigger(ctx, tokens)
		if err != nil {
			return nil, err
		}

		entries = append(entries, e...)
	}

	return entries, nil
}
