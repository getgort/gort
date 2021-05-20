package rest

// User is a data struct used to exchange data between gortctl and the REST service.
type User struct {
	Email    string `json:"email,omitempty"`
	FullName string `json:"fullname,omitempty"`
	Password string `json:"password,omitempty"`
	Username string `json:"username,omitempty"`
}
