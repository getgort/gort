package rest

// User is a data struct used to exchange data between cogctl and the REST service.
type User struct {
	Email    string `json:"email,omitempty"`
	FullName string `json:"first-name,omitempty"`
	Password string `json:"password,omitempty"`
	Username string `json:"username,omitempty"`
}
