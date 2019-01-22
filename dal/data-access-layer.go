package dal

import (
	"crypto/rand"
	b64 "encoding/base64"
	"time"

	"github.com/clockworksoul/cog2/data/rest"
	"golang.org/x/crypto/bcrypt"
)

// DataAccess represents a common DataAccessObject, backed either by a
// database or an in-memory datastore.
type DataAccess interface {
	Initialize() error
	GroupAddUser(string, string) error
	GroupCreate(rest.Group) error
	GroupDelete(string) error
	GroupExists(string) (bool, error)
	GroupGet(string) (rest.Group, error)
	GroupGrantRole() error
	GroupList() ([]rest.Group, error)
	GroupRemoveUser(string, string) error
	GroupRevokeRole() error
	GroupUpdate(rest.Group) error

	UserAuthenticate(string, string) (bool, error)
	UserCreate(rest.User) error
	UserDelete(string) error
	UserExists(string) (bool, error)
	UserGet(string) (rest.User, error)
	UserList() ([]rest.User, error)
	UserUpdate(rest.User) error

	TokenEvaluate(token string) bool
	TokenGenerate(user string, duration time.Duration) (Token, error)
	TokenInvalidate(token string) error
	TokenRetrieveByUser(user string) (Token, error)
	TokenRetrieveByToken(token string) (Token, error)
}

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

// CompareHashAndPassword receives a plaintext password and its hash, and
// returns true if they match.
func CompareHashAndPassword(hashedPassword string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

// GenerateRandomToken generates a random character token.
func GenerateRandomToken(length int) (string, error) {
	byteCount := (length * 3) / 4
	bytes := make([]byte, byteCount)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	sEnc := b64.StdEncoding.EncodeToString(bytes)

	return sEnc, nil
}

// HashPassword receives a plaintext password and returns its hashed equivalent.
func HashPassword(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
