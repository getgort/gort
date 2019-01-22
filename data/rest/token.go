package rest

import "time"

// Token contains all of the metadata for an access token.
type Token struct {
	Duration   time.Duration `json:"-"`
	Token      string        `json:",omitempty"`
	User       string        `json:",omitempty"`
	ValidFrom  time.Time     `json:",omitempty"`
	ValidUntil time.Time     `json:",omitempty"`
}

// IsExpired returns true if the token has expired.
func (t Token) IsExpired() bool {
	return !t.ValidUntil.After(time.Now())
}
