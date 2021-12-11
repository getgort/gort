package data

import (
	"fmt"
	"strings"
)

const (
	LayerBundle ConfigurationLayer = "bundle"
	LayerRoom   ConfigurationLayer = "room"
	LayerGroup  ConfigurationLayer = "group"
	LayerUser   ConfigurationLayer = "user"
)

type ConfigurationLayer string

func (c ConfigurationLayer) Validate() error {
	s := ConfigurationLayer(strings.ToLower(string(c)))
	if s != LayerBundle && s != LayerRoom && s != LayerGroup && s != LayerUser {
		return fmt.Errorf("dynamic configuration layers must be one of: %v",
			[]ConfigurationLayer{LayerBundle, LayerRoom, LayerGroup, LayerUser})
	}
	return nil
}

type DynamicConfiguration struct {
	// Bundle is the bundle this layer is associated with.
	Bundle string

	// Layer is the "layer". Must be one of the Layer* constants.
	Layer ConfigurationLayer

	// Owner is the entity that own this. If the layer is "room", this is the
	// name of the room. If it's "group", it's the nam eof the group, etc.
	Owner string

	// Key is the key component of the key-value pair. Should be an all-lower
	// general name that's unique for the same values of Layer and Owner.
	Key string

	// Value is the value itself. Can be any valid string.
	// TODO(mtitmus) Should there be a length limit? If so, it should still be
	// pretty big.
	Value string

	// Secret is true if this value isn't allowed to be viewed by others.
	// Note that encryption is part of the storage backend implementation: this
	// value is unrelated to that.
	Secret bool
}
