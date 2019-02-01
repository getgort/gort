package postgres

import (
	"github.com/clockworksoul/cog2/data"
	"github.com/clockworksoul/cog2/dataaccess/errs"
)

// BundleCreate TBD
func (da PostgresDataAccess) BundleCreate(bundle data.Bundle) error {
	return errs.ErrNotImplemented
}

// BundleDelete TBD
func (da PostgresDataAccess) BundleDelete(bundlename string) error {
	return errs.ErrNotImplemented
}

// BundleExists TBD
func (da PostgresDataAccess) BundleExists(bundlename string) (bool, error) {
	return false, errs.ErrNotImplemented
}

// BundleGet TBD
func (da PostgresDataAccess) BundleGet(bundlename string) (data.Bundle, error) {
	return data.Bundle{}, errs.ErrNotImplemented
}

// BundleList TBD
func (da PostgresDataAccess) BundleList() ([]data.Bundle, error) {
	return []data.Bundle{}, errs.ErrNotImplemented
}

// BundleUpdate TBD
func (da PostgresDataAccess) BundleUpdate(bundle data.Bundle) error {
	return errs.ErrNotImplemented
}
