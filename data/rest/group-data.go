package rest

// Group is a data struct used to exchange data between cogctl and the REST service.
type Group struct {
	Name  string `json:"name,omitempty"`
	Users []User `json:"users,omitempty"`
}
