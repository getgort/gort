package data

import (
	"fmt"
	"strings"
	"time"
)

// BundleInfo wraps a minimal amount of data abount a bundle.
type BundleInfo struct {
	Name           string
	Versions       []string
	Enabled        bool
	EnabledVersion Bundle
}

// Bundle represents a bundle as defined in the "bundles" section of the
// config.
type Bundle struct {
	CogBundleVersion int                      `yaml:"cog_bundle_version,omitempty" json:"cog_bundle_version,omitempty"`
	Name             string                   `yaml:",omitempty" json:"name,omitempty"`
	Version          string                   `yaml:",omitempty" json:"version,omitempty"`
	Enabled          bool                     `yaml:",omitempty" json:"enabled"`
	Author           string                   `yaml:",omitempty" json:"author,omitempty"`
	Homepage         string                   `yaml:",omitempty" json:"homepage,omitempty"`
	Description      string                   `yaml:",omitempty" json:"description,omitempty"`
	InstalledOn      time.Time                `yaml:"-" json:"installed_on,omitempty"`
	InstalledBy      string                   `yaml:",omitempty" json:"installed_by,omitempty"`
	LongDescription  string                   `yaml:"long_description,omitempty" json:"long_description,omitempty"`
	Docker           BundleDocker             `yaml:",omitempty" json:"docker,omitempty"`
	Permissions      []string                 `yaml:",omitempty" json:"permissions,omitempty"`
	Commands         map[string]BundleCommand `yaml:",omitempty" json:"commands,omitempty"`
}

// BundleDocker represents the "bundles/docker" subsection of the config doc
type BundleDocker struct {
	Image string `yaml:",omitempty" json:"image,omitempty"`
	Tag   string `yaml:",omitempty" json:"tag,omitempty"`
}

// BundleCommand represents a bundle command, as defined in the "bundles/commands"
// section of the config.
type BundleCommand struct {
	Description string   `yaml:",omitempty" json:"description,omitempty"`
	Executable  string   `yaml:",omitempty" json:"executable,omitempty"`
	Name        string   `yaml:"-" json:"-"`
	Rules       []string `yaml:",omitempty" json:"rules,omitempty"`
}

// CommandEntry conveniently wraps a bundle and one command within that bundle.
type CommandEntry struct {
	Bundle  Bundle
	Command BundleCommand
}

// CommandRequest represents a user command request as triggered in (probably)
// a chat provider.
type CommandRequest struct {
	CommandEntry
	Adapter    string   // The name of the adapter this request originated from.
	ChannelID  string   // The channel that the request originated in.
	UserID     string   // The ID of the user making this request.
	Parameters []string // Tokenized command parameters
}

// CommandString is a convenience method that outputs the normalized command
// string, more or less as the user typed it.
func (r CommandRequest) CommandString() string {
	return fmt.Sprintf(
		"%s:%s %s",
		r.Bundle.Name,
		r.Command.Name,
		strings.Join(r.Parameters, " "))
}

// CommandResponse is returned by a relay to indicate that a command has been executed.
// It includes the original CommandRequest, the command's exit status code, and
// the commands entire stdout as a slice of lines. Title can be used to build
// a user output message, and generally contains a short description of the result.
//
// TODO Add a request ID that correcponds with the request, so that we can more
// directly link it back to its user and adapter of origin.
type CommandResponse struct {
	Command CommandRequest
	Status  int64
	Title   string   // Command Error
	Output  []string // Contents of the commands stdout.
	Error   error
}
