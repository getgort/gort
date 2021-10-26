/*
 * Copyright 2021 The Gort Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/getgort/gort/data/rest"
	gerrs "github.com/getgort/gort/errors"
)

// Authenticate requests a new authentication token from the Gort controller.
// If a valid token already exists it will be automatically invalidated if
// this call is successful.
func (c *GortClient) Authenticate() (rest.Token, error) {
	// If the GORT_SERVICE_TOKEN envvar is set, use that first.
	if te, exists := os.LookupEnv("GORT_SERVICE_TOKEN"); exists {
		token := rest.Token{
			Token:      te,
			ValidFrom:  time.Now(),
			ValidUntil: time.Now().Add(10 * time.Second),
		}

		return token, nil
	}

	endpointURL := fmt.Sprintf("%s/v2/authenticate", c.profile.URL)

	postBytes, err := json.Marshal(c.profile.User())
	if err != nil {
		return rest.Token{}, gerrs.Wrap(gerrs.ErrMarshal, err)
	}

	resp, err := http.Post(endpointURL, "application/json", bytes.NewBuffer(postBytes))
	switch {
	case err == nil:
	case strings.Contains(err.Error(), "certificate"):
		return rest.Token{}, fmt.Errorf("self-signed certificate detected: use --allow-insecure to proceed (not recommended)")
	default:
		return rest.Token{}, gerrs.Wrap(ErrConnectionFailed, err)
	}

	if resp.StatusCode != http.StatusOK {
		bytes, _ := ioutil.ReadAll(resp.Body)
		return rest.Token{}, fmt.Errorf(string(bytes))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return rest.Token{}, gerrs.Wrap(ErrResponseReadFailure, err)
	}

	token := rest.Token{}
	err = json.Unmarshal(body, &token)
	if err != nil {
		return rest.Token{}, gerrs.Wrap(gerrs.ErrUnmarshal, err)
	}

	// Save the token to disk
	file, err := c.getGortTokenFilename()
	if err != nil {
		return token, gerrs.Wrap(gerrs.ErrIO, err)
	}

	f, err := os.Create(file)
	if err != nil {
		return token, gerrs.Wrap(gerrs.ErrIO, err)
	}
	defer f.Close()

	_, err = f.Write(body)
	if err != nil {
		return token, gerrs.Wrap(gerrs.ErrIO, err)
	}

	return token, nil
}

// Authenticated looks for any cached tokens associated with the current
// server. Returns false if no tokens exist or tokens are expired.
func (c *GortClient) Authenticated() (bool, error) {
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
func (c *GortClient) Bootstrap() (rest.User, error) {
	endpointURL := fmt.Sprintf("%s/v2/bootstrap", c.profile.URL)

	// Get profile data so we can update it afterwards
	profile, err := LoadClientProfile()
	if err != nil {
		return rest.User{}, err
	}

	if _, exists := profile.Profiles[c.profile.Name]; exists {
		return rest.User{}, fmt.Errorf("profile %s already exists", c.profile.Name)
	}

	postBytes, err := json.Marshal(rest.User{})
	if err != nil {
		return rest.User{}, gerrs.Wrap(gerrs.ErrMarshal, err)
	}

	resp, err := c.client.Post(endpointURL, "application/json", bytes.NewBuffer(postBytes))
	switch {
	case err == nil:
	case strings.Contains(err.Error(), "certificate"):
		return rest.User{}, fmt.Errorf("self-signed certificate detected: use --allow-insecure to proceed (not recommended)")
	default:
		return rest.User{}, gerrs.Wrap(ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK: // Everything is swell.
	case http.StatusConflict:
		err := fmt.Errorf("server %s has already been bootstrapped", c.profile.URL)
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
		return rest.User{}, gerrs.Wrap(gerrs.ErrIO, err)
	}

	user := rest.User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		return rest.User{}, gerrs.Wrap(gerrs.ErrUnmarshal, err)
	}

	// Update the client profile file
	entry := c.profile
	entry.Password = user.Password
	entry.Username = user.Username

	if profile.Defaults.Profile == "" {
		profile.Defaults.Profile = entry.Name
	}

	profile.Profiles[entry.Name] = entry
	err = SaveClientProfile(profile)
	if err != nil {
		return user, err
	}

	return user, nil
}

// Token is just a wrapper around a call to Authenticated() followed by a
// call to Authenticate() if false.
func (c *GortClient) Token() (rest.Token, error) {
	authed, err := c.Authenticated()
	if err != nil {
		return rest.Token{}, err
	}

	if authed {
		return *c.token, nil
	}

	return c.Authenticate()
}
