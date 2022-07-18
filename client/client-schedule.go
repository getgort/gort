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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/getgort/gort/data/rest"
)

// ScheduleCreate schedules a new command to run periodically.
func (c *GortClient) ScheduleCreate(req rest.ScheduleRequest) (int64, error) {
	url := fmt.Sprintf("%s/v2/schedules", c.profile.URL.String())

	body, err := json.Marshal(req)
	if err != nil {
		return 0, err
	}

	resp, err := c.doRequest("PUT", url, body)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, getResponseError(resp)
	}

	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	id, err := strconv.ParseInt(string(r), 0, 64)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (c *GortClient) SchedulesGet() ([]rest.ScheduleInfo, error) {
	url := fmt.Sprintf("%s/v2/schedules", c.profile.URL.String())

	resp, err := c.doRequest("GET", url, []byte{})
	if err != nil {
		return []rest.ScheduleInfo{}, err
	}
	defer resp.Body.Close()

	var info []rest.ScheduleInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	return info, err
}

func (c *GortClient) ScheduleDelete(id int64) error {
	url := fmt.Sprintf("%s/v2/schedules/%d", c.profile.URL.String(), id)

	resp, err := c.doRequest("DELETE", url, []byte{})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
