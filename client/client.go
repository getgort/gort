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
	"strings"

	"github.com/clockworksoul/cog2/data/rest"
	homedir "github.com/mitchellh/go-homedir"
)

// CogClient comments to be written...
type CogClient struct {
	profile ProfileEntry
	token   *rest.Token
}

// Connect creates and returns a configured instance of the client for the
// specified host. An empty string will use the default profile. If the
// requested profile doesn't exist, an empty ProfileEntry is returned.
func Connect(profileName string) (*CogClient, error) {
	var entry ProfileEntry

	// Load the profiles file
	profile, err := loadClientProfile()
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("no such profile: %s", profileName)
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

func getResponseError(resp *http.Response) error {
	bytes, _ := ioutil.ReadAll(resp.Body)
	customError := strings.TrimSpace(string(bytes))
	return fmt.Errorf("%d %s", resp.StatusCode, customError)
}

func (c *CogClient) doRequest(method string, url string, body []byte) (*http.Response, error) {
	token, err := c.Token()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	req.Header.Add("X-Session-Token", token.Token)

	client := &http.Client{}
	return client.Do(req)
}

// getCogTokenFilename finds and returns the full-qualified filename for this
// host's token file, stored in the $HOME/.cog/tokens directory.
func (c *CogClient) getCogTokenFilename() (string, error) {
	cogDir, err := getCogTokenDir()
	if err != nil {
		return "", err
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
	err = json.Unmarshal(bytes, &token)
	if err != nil {
		return token, err
	}

	return token, nil
}

// parseHostURL receives a host url string and returns a pointer *url.URL
// pointer.  Unlike url.Parse(), this function will assume a scheme of "http"
// if a scheme is not specified.
func parseHostURL(serverURLArg string) (*url.URL, error) {
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
