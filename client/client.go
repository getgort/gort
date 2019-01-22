package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"

	"github.com/clockworksoul/cog2/data/rest"
)

// CogClient COMMENT ME
type CogClient struct {
	host  *url.URL
	token *rest.Token
}

// Connect COMMENT ME
func Connect(host string) (*CogClient, error) {
	hostURL, err := ParseHostURL(host)
	if err != nil {
		return nil, err
	}

	client := &CogClient{host: hostURL, token: nil}

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
func (c *CogClient) Authenticate() (rest.Token, error) {
	return rest.Token{}, fmt.Errorf("not yet implemented")
}

// Authenticated looks for any cached tokens associated with the current
// server. Returns false if no tokens exist or tokens are expired.
func (c *CogClient) Authenticated() bool {
	// If the token var isn't set, look in the users cache.
	if c.token == nil {
		c.token, _ = c.loadHostToken()
	}

	return c.token != nil && !c.token.IsExpired()
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
	if c.Authenticated() {
		return *c.token, nil
	}

	return c.Authenticate()
}

// IsTLS returns true iff the host URL scheme is "https".
func (c *CogClient) IsTLS() bool {
	return c.host.Scheme == "https"
}

func (c *CogClient) loadHostToken() (*rest.Token, error) {
	return nil, fmt.Errorf("not yet implemented")
}
