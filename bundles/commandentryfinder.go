package bundles

import "github.com/getgort/gort/data"

// FindCommandEntry is used to find the enabled commands with the provided
// bundle and command names. If either is empty, it is treated as a wildcard.
// Importantly, this must only return ENABLED commands!
type CommandEntryFinder interface {
	FindCommandEntry(bundle, command string) ([]data.CommandEntry, error)
}
