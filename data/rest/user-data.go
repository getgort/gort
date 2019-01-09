package rest

// User is a data struct used to exchange data between cogctl and the REST service.
type User struct {
	Email     string `json:"email,omitempty"`
	FirstName string `json:"first-name,omitempty"`
	LastName  string `json:"last-name,omitempty"`
	Password  string `json:"password,omitempty"`
	Username  string `json:"username,omitempty"`
}
