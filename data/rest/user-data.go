package rest

type User struct {
	Email    string `json:"email,omitempty"`
	FullName string `json:"full-name,omitempty"`
	Password string `json:"password,omitempty"`
	Username string `json:"username,omitempty"`
}
