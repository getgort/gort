package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/clockworksoul/cog2/data/rest"
	cogerr "github.com/clockworksoul/cog2/errors"
	homedir "github.com/mitchellh/go-homedir"
)

var (
	// ErrBadProfile indicates an invalid or missing client profile.
	ErrBadProfile = errors.New("invalid or missing client profile")

	// ErrBadRequest indicates that a request could not be constructed.
	ErrBadRequest = errors.New("request could not be constructed")

	// ErrConnectionFailed is a failure for a client to connect to the Cog service.
	ErrConnectionFailed = errors.New("failure to connect to the Cog service")

	// ErrResourceExists is returned if a client tries to put a resource that
	// already exists.
	ErrResourceExists = errors.New("resource already exists")

	// ErrResourceNotFound is returned if a client tries to get or update a
	// resource that doesn't exist.
	ErrResourceNotFound = errors.New("resource doesn't exist")

	// ErrResponseReadFailure indicates an error in reading a server response.
	ErrResponseReadFailure = errors.New("error reading a server response")

	// ErrURLFormat indicates badly formatted URL.
	ErrURLFormat = errors.New("invalid URL format")
)

// CogClient comments to be written...
type CogClient struct {
	profile ProfileEntry
	token   *rest.Token
}

// Error is an error implementation that represents either a a non-2XX
// response from the server, or a failure to connect to the server (in which
// case Status() will return 0).
type Error struct {
	error
	profile ProfileEntry
	status  uint
}

// Error returns the error message for this error.
func (c Error) Error() string {
	return c.error.Error()
}

// Profile returns the active profile entry for the client that returned
// this error.
func (c Error) Profile() ProfileEntry {
	return c.profile
}

// Status returns the HTTP status code provided by the server. A status of
// 0 indicates that the client failed to connect entirely.
func (c Error) Status() uint {
	return c.status
}

// Connect creates and returns a configured instance of the client for the
// specified host. An empty string will use the default profile. If the
// requested profile doesn't exist, an empty ProfileEntry is returned.
func Connect(profileName string) (*CogClient, error) {
	var entry ProfileEntry

	// Load the profiles file
	profile, err := loadClientProfile()
	if err != nil {
		return nil, cogerr.Wrap(ErrBadProfile, err)
	}

	// Find the desired profile entry
	if profileName == "" {
		entry = profile.Default()
	} else {
		ok := false
		entry, ok = profile.Profiles[profileName]

		if ok {
			entry.Name = profileName
		}
	}

	if entry.Name == "" {
		return nil, ErrBadProfile
	}

	return &CogClient{profile: entry}, nil
}

// ConnectWithNewProfile generates a connection using the supplied profile
// entry data.
func ConnectWithNewProfile(entry ProfileEntry) (*CogClient, error) {
	url, err := parseHostURL(entry.URLString)
	if err != nil {
		return nil, err
	}

	entry.URL = url
	entry.URLString = url.String()

	if entry.Name == "" {
		entry.Name = url.Hostname()
	}

	return &CogClient{profile: entry}, nil
}

func (c *CogClient) doRequest(method string, url string, body []byte) (*http.Response, error) {
	token, err := c.Token()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, cogerr.Wrap(ErrBadRequest, err)
	}
	req.Header.Add("X-Session-Token", token.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, cogerr.Wrap(ErrConnectionFailed, err)
	}

	return resp, err
}

// getCogTokenFilename finds and returns the full-qualified filename for this
// host's token file, stored in the $HOME/.cog/tokens directory.
func (c *CogClient) getCogTokenFilename() (string, error) {
	cogDir, err := getCogTokenDir()
	if err != nil {
		return "", cogerr.Wrap(cogerr.ErrIO, err)
	}

	url := c.profile.URL
	tokenFileName := fmt.Sprintf("%s/%s_%s", cogDir, url.Hostname(), url.Port())

	return tokenFileName, nil
}

// loadHostToken attempts to load an existing token from a file. If the token
// file exists, a filled Token{} is returned; an empty Token{} is it doesn't.
// An error is only returned is there's an underlying error.
func (c *CogClient) loadHostToken() (rest.Token, error) {
	tokenFileName, err := c.getCogTokenFilename()
	if err != nil {
		return rest.Token{}, cogerr.Wrap(cogerr.ErrIO, err)
	}

	// File doesn't exist. Not an error.
	if _, err := os.Stat(tokenFileName); err != nil {
		return rest.Token{}, nil
	}

	bytes, err := ioutil.ReadFile(tokenFileName)
	if err != nil {
		return rest.Token{}, cogerr.Wrap(cogerr.ErrIO, err)
	}

	token := rest.Token{}
	err = json.Unmarshal(bytes, &token)
	if err != nil {
		return token, cogerr.Wrap(cogerr.ErrUnmarshal, err)
	}

	return token, nil
}

// getCogConfigDir finds the users $HOME/.cog directory, creating it if it
// doesn't exist.
func getCogConfigDir() (string, error) {
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
		merr := os.Mkdir(cogDir, 0700)
		if merr != nil {
			return "", merr
		}
	}

	return cogDir, nil
}

// getCogConfigDir finds the users $HOME/.cog/tokens directory, creating it if
// it doesn't exist.
func getCogTokenDir() (string, error) {
	cogDir, err := getCogConfigDir()
	if err != nil {
		return "", err
	}

	tokenDir := cogDir + "/tokens"

	if tokenDirInfo, err := os.Stat(tokenDir); err == nil {
		if !tokenDirInfo.IsDir() {
			return "", fmt.Errorf("%s exists but is not a directory", tokenDir)
		}
	} else if os.IsNotExist(err) {
		merr := os.Mkdir(tokenDir, 0700)
		if merr != nil {
			return "", merr
		}
	}

	return tokenDir, nil
}

// getResponseError receives an http.Response pointer and returns an Error
// from its status message and code.
func getResponseError(resp *http.Response) Error {
	bytes, _ := ioutil.ReadAll(resp.Body)
	status := strings.TrimSpace(string(bytes))
	code := uint(resp.StatusCode)

	if status == "" {
		status = resp.Status
	}

	if strings.HasPrefix(status, fmt.Sprintf("%d ", code)) {
		status = status[4:]
	}

	return Error{error: errors.New(status), status: code}
}

// parseHostURL receives a host url string and returns a pointer *url.URL
// pointer.  Unlike url.Parse(), this function will assume a scheme of "http"
// if a scheme is not specified.
func parseHostURL(serverURLArg string) (*url.URL, error) {
	serverURLString := serverURLArg

	// Does the URL have a prefix? If not, assume 'http://'
	matches, err := regexp.MatchString("^[a-z0-9]+://.*", serverURLString)
	if err != nil {
		return nil, cogerr.Wrap(cogerr.ErrIO, err)
	}
	if !matches {
		serverURLString = "http://" + serverURLString
	}

	// Parse the resulting URL
	serverURL, err := url.Parse(serverURLString)
	if err != nil {
		return nil, cogerr.Wrap(ErrURLFormat, err)
	}

	return serverURL, nil
}
