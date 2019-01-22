package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/clockworksoul/cog2/data/rest"
	homedir "github.com/mitchellh/go-homedir"
)

// CogClient COMMENT ME
type CogClient struct {
	host  *url.URL
	token *rest.Token
	user  *rest.User
}

// Connect COMMENT ME
func Connect(host string) (*CogClient, error) {
	hostURL, err := ParseHostURL(host)
	if err != nil {
		return nil, err
	}

	// TODO: Load the user if it exists in the config
	user := rest.User{
		Username: "admin",
		Password: "password",
	}

	client := &CogClient{
		host:  hostURL,
		user:  &user,
		token: nil,
	}

	return client, nil
}

// ParseHostURL receives a host url string and returns a pointer *url.URL
// pointer.  Unlike url.Parse(), this function will assume a scheme of "http"
// if a scheme is not specified.
func ParseHostURL(serverURLArg string) (*url.URL, error) {
	serverURLString := serverURLArg

	// Does the URL have a prefix? If not, assume 'http://'
	matches, err := regexp.MatchString("^[a-z0-9]+://.*", serverURLString)
	if err != nil {
		return nil, err
	}
	if !matches {
		serverURLString = "http://" + serverURLString
	}

	// Parse the resulting URL
	serverURL, err := url.Parse(serverURLString)
	if err != nil {
		return nil, err
	}

	return serverURL, nil
}

// Authenticate requests a new authentication token from the Cog service.
// If a valid token already exists it will be automatically invalidated if
// this call is successful.
func (c *CogClient) Authenticate(user rest.User) (rest.Token, error) {
	endpointURL := fmt.Sprintf("%s/v2/authenticate", c.host.String())

	postBytes, err := json.Marshal(user)
	if err != nil {
		return rest.Token{}, err
	}

	resp, err := http.Post(endpointURL, "application/json", bytes.NewBuffer(postBytes))
	if err != nil {
		return rest.Token{}, err
	}

	if resp.StatusCode != http.StatusOK {
		bytes, _ := ioutil.ReadAll(resp.Body)
		return rest.Token{}, fmt.Errorf(string(bytes))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return rest.Token{}, err
	}

	token := rest.Token{}
	err = json.Unmarshal(body, &token)
	if err != nil {
		return rest.Token{}, err
	}

	return token, nil
}

// Authenticated looks for any cached tokens associated with the current
// server. Returns false if no tokens exist or tokens are expired.
func (c *CogClient) Authenticated() (bool, error) {
	if c.token != nil {
		return !c.token.IsExpired(), nil
	}

	// If the token var isn't set, look in the users cache.
	token, err := c.loadHostToken()
	if err != nil {
		return false, err
	}

	// Empty token means no error, but no token. :(
	if token.Token == "" {
		return false, nil
	}

	// We found a token! Keep it.
	c.token = &token
	return !c.token.IsExpired(), nil
}

// Bootstrap calls the POST /v2/bootstrap endpoint.
func (c *CogClient) Bootstrap(user rest.User) (rest.User, error) {
	endpointURL := fmt.Sprintf("%s/v2/bootstrap", c.host.String())

	postBytes, err := json.Marshal(user)
	if err != nil {
		return rest.User{}, err
	}

	resp, err := http.Post(endpointURL, "application/json", bytes.NewBuffer(postBytes))
	if err != nil {
		return rest.User{}, err
	}

	switch resp.StatusCode {
	case http.StatusOK: // Everything is swell.
	case http.StatusConflict:
		err := fmt.Errorf("server %s has already been bootstrapped", c.host.String())
		return rest.User{}, err
	case http.StatusInternalServerError:
		err := fmt.Errorf("internal server error; check the server logs for details")
		return rest.User{}, err
	default:
		bytes, _ := ioutil.ReadAll(resp.Body)
		return rest.User{}, fmt.Errorf(string(bytes))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return rest.User{}, err
	}

	// Re-using "user" instance. Sorry.
	err = json.Unmarshal(body, &user)
	if err != nil {
		return rest.User{}, err
	}

	return user, nil
}

// Host returns the URL that this client is interacting with.
func (c *CogClient) Host() *url.URL {
	return c.host
}

// Token is just a wrapper around a call to Authenticated() followed by a
// call to Authenticate() if false.
func (c *CogClient) Token() (rest.Token, error) {
	authed, err := c.Authenticated()
	if err != nil {
		return rest.Token{}, err
	}

	if authed {
		return *c.token, nil
	}

	return c.Authenticate(*c.user)
}

// IsTLS returns true iff the host URL scheme is "https".
func (c *CogClient) IsTLS() bool {
	return c.host.Scheme == "https"
}

func (c *CogClient) getCogConfigDir() (string, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	cogDir := homeDir + "/.cog"

	if cogDirInfo, err := os.Stat(cogDir); err == nil {
		if !cogDirInfo.IsDir() {
			return "", fmt.Errorf("%s exists but is not a directory", cogDir)
		}
	} else if os.IsNotExist(err) {
		merr := os.Mkdir(cogDir, 0500)
		if merr != nil {
			return "", merr
		}
	}

	return cogDir, nil
}

func (c *CogClient) getCogTokenDir() (string, error) {
	cogDir, err := c.getCogConfigDir()
	if err != nil {
		return "", err
	}

	tokenDir := cogDir + "/tokens"

	if tokenDirInfo, err := os.Stat(tokenDir); err == nil {
		if !tokenDirInfo.IsDir() {
			return "", fmt.Errorf("%s exists but is not a directory", tokenDir)
		}
	} else if os.IsNotExist(err) {
		merr := os.Mkdir(tokenDir, 0500)
		if merr != nil {
			return "", merr
		}
	}

	return "", nil
}

func (c *CogClient) getCogTokenFilename() (string, error) {
	cogDir, err := c.getCogTokenDir()
	if err != nil {
		return "", err
	}

	tokenFileName := fmt.Sprintf("%s/%s_%s", cogDir, c.host.Hostname(), c.host.Port())

	return tokenFileName, nil
}

// loadHostToken attempts to load an existing token from a file. If the token
// file exists, a filled Token{} is returned; an empty Token{} is it doesn't.
// An error is only returned is there's an underlying error.
func (c *CogClient) loadHostToken() (rest.Token, error) {
	tokenFileName, err := c.getCogTokenFilename()
	if err != nil {
		return rest.Token{}, err
	}

	// File doesn't exist. Not an error.
	if _, err := os.Stat(tokenFileName); err != nil {
		return rest.Token{}, nil
	}

	bytes, err := ioutil.ReadFile(tokenFileName)
	if err != nil {
		return rest.Token{}, err
	}

	token := rest.Token{}
	err = json.Unmarshal(bytes, token)
	if err != nil {
		return token, err
	}

	return token, nil
}
