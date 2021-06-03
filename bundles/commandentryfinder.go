package bundles

import "github.com/getgort/gort/data"

type CommandEntryFinder interface {
	FindCommandEntry(bundle, command string) ([]data.CommandEntry, error)
}
