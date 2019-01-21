package memory

import (
	"github.com/clockworksoul/cog2/dal"
	"github.com/clockworksoul/cog2/data/rest"
)

// InMemoryDataAccess is an entirely in-memory representation of a data access layer.
// Great for testing and development. Terrible for production.
type InMemoryDataAccess struct {
	dal.DataAccess

	groups map[string]*rest.Group
	users  map[string]*rest.User
}

// NewInMemoryDataAccess returns a new InMemoryDataAccess instance.
func NewInMemoryDataAccess() dal.DataAccess {
	da := InMemoryDataAccess{
		groups: make(map[string]*rest.Group),
		users:  make(map[string]*rest.User),
	}

	return da
}

// Initialize initializes an InMemoryDataAccess instance.
func (da InMemoryDataAccess) Initialize() error {
	return nil
}
