package memory

import (
	"github.com/clockworksoul/cog2/dal"
	"github.com/clockworksoul/cog2/data/rest"
)

func NewInMemoryDataAccess() dal.DataAccess {
	da := InMemoryDataAccess{
		groups: make(map[string]*rest.Group),
		users:  make(map[string]*rest.User),
	}

	return da
}

type InMemoryDataAccess struct {
	dal.DataAccess

	groups map[string]*rest.Group
	users  map[string]*rest.User
}
