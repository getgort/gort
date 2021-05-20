package adapter

// UserInfo contains the basic information for a single user in any chat provider.
type UserInfo struct {
	ID                    string
	Name                  string
	DisplayName           string
	DisplayNameNormalized string
	Email                 string
	FirstName             string
	LastName              string
	RealName              string
	RealNameNormalized    string
}
